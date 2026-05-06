package game

import (
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
)

func TestCombatResolver_ResolveAttack(t *testing.T) {
	resolver := NewCombatResolver()

	tests := []struct {
		name        string
		attacker    domain.PlayerState
		target      domain.PlayerState
		attackRoll  int
		wantHit     bool
		wantCrit    bool
		wantCritFail bool
	}{
		{
			name:       "normal hit",
			attacker:   domain.PlayerState{CharacterID: "attacker", AC: 15},
			target:     domain.PlayerState{CharacterID: "target", AC: 12},
			attackRoll: 15, // 15 >= 12 AC = hit
			wantHit:    true,
			wantCrit:   false,
			wantCritFail: false,
		},
		{
			name:       "normal miss",
			attacker:   domain.PlayerState{CharacterID: "attacker", AC: 15},
			target:     domain.PlayerState{CharacterID: "target", AC: 16},
			attackRoll: 10, // 10 < 16 AC = miss
			wantHit:    false,
			wantCrit:   false,
			wantCritFail: false,
		},
		{
			name:       "critical hit",
			attacker:   domain.PlayerState{CharacterID: "attacker", AC: 15},
			target:     domain.PlayerState{CharacterID: "target", AC: 20},
			attackRoll: 20, // natural 20 = crit (always hits)
			wantHit:    true,
			wantCrit:   true,
			wantCritFail: false,
		},
		{
			name:       "critical fail",
			attacker:   domain.PlayerState{CharacterID: "attacker", AC: 15},
			target:     domain.PlayerState{CharacterID: "target", AC: 5},
			attackRoll: 1, // natural 1 = miss regardless
			wantHit:    false,
			wantCrit:   false,
			wantCritFail: true,
		},
		{
			name:       "exact AC - hit",
			attacker:   domain.PlayerState{CharacterID: "attacker", AC: 15},
			target:     domain.PlayerState{CharacterID: "target", AC: 12},
			attackRoll: 12, // 12 == 12 AC = hit
			wantHit:    true,
			wantCrit:   false,
			wantCritFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolver.ResolveAttack(&tt.attacker, &tt.target, tt.attackRoll, 0)

			if result.Hit != tt.wantHit {
				t.Errorf("ResolveAttack() hit = %v, want %v", result.Hit, tt.wantHit)
			}
			if result.CriticalHit != tt.wantCrit {
				t.Errorf("ResolveAttack() crit = %v, want %v", result.CriticalHit, tt.wantCrit)
			}
			if result.CriticalFail != tt.wantCritFail {
				t.Errorf("ResolveAttack() critFail = %v, want %v", result.CriticalFail, tt.wantCritFail)
			}
		})
	}
}

func TestCombatResolver_CalculateDamage(t *testing.T) {
	resolver := NewCombatResolver()

	tests := []struct {
		name       string
		dice       string
		isCritical bool
		wantMin    int
		wantMax    int
	}{
		{
			name:       "normal damage 1d8",
			dice:       "1d8",
			isCritical: false,
			wantMin:    1,
			wantMax:    8,
		},
		{
			name:       "critical damage 1d8",
			dice:       "1d8",
			isCritical: true,
			wantMin:    2,
			wantMax:    16,
		},
		{
			name:       "normal damage 2d6",
			dice:       "2d6",
			isCritical: false,
			wantMin:    2,
			wantMax:    12,
		},
		{
			name:       "critical damage 2d6",
			dice:       "2d6",
			isCritical: true,
			wantMin:    4,
			wantMax:    24,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			damage, err := resolver.CalculateDamage(tt.dice, tt.isCritical)
			if err != nil {
				t.Fatalf("CalculateDamage() error: %v", err)
			}

			if damage < tt.wantMin || damage > tt.wantMax {
				t.Errorf("CalculateDamage() = %d, want between %d and %d", damage, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestCombatResolver_CalculateInitiative(t *testing.T) {
	resolver := NewCombatResolver()

	actors := []InitiativeActor{
		{CharacterID: "char-1", DEXModifier: 2},
		{CharacterID: "char-2", DEXModifier: 0},
		{CharacterID: "char-3", DEXModifier: 4},
	}

	// Roll initiative (we can't predict the exact order because of dice rolls,
	// but we can verify the structure)
	order := resolver.CalculateInitiative(actors)

	if len(order) != len(actors) {
		t.Errorf("CalculateInitiative() returned %d actors, want %d", len(order), len(actors))
	}

	// Verify all actors are in the order
	actorIDs := make(map[string]bool)
	for _, actor := range order {
		actorIDs[actor.CharacterID] = true
	}
	for _, actor := range actors {
		if !actorIDs[actor.CharacterID] {
			t.Errorf("CalculateInitiative() missing actor %s", actor.CharacterID)
		}
	}
}

func TestCombatResolver_ApplyDamageToPlayer(t *testing.T) {
	resolver := NewCombatResolver()

	tests := []struct {
		name        string
		player      domain.PlayerState
		damage      int
		wantHP      int
		wantAlive   bool
		wantTempHP  int
	}{
		{
			name:       "simple damage",
			player:     domain.PlayerState{CurrentHP: 10, MaxHP: 10, TempHP: 0},
			damage:     3,
			wantHP:     7,
			wantAlive:  true,
			wantTempHP: 0,
		},
		{
			name:       "damage with temp HP",
			player:     domain.PlayerState{CurrentHP: 10, MaxHP: 10, TempHP: 5},
			damage:     3,
			wantHP:     10,
			wantAlive:  true,
			wantTempHP: 2,
		},
		{
			name:       "lethal damage",
			player:     domain.PlayerState{CurrentHP: 5, MaxHP: 10, TempHP: 0},
			damage:     10,
			wantHP:     0,
			wantAlive:  false,
			wantTempHP: 0,
		},
		{
			name:       "zero damage",
			player:     domain.PlayerState{CurrentHP: 10, MaxHP: 10, TempHP: 0},
			damage:     0,
			wantHP:     10,
			wantAlive:  true,
			wantTempHP: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := tt.player
			resolver.ApplyDamageToPlayer(&player, tt.damage)

			if player.CurrentHP != tt.wantHP {
				t.Errorf("ApplyDamageToPlayer() HP = %d, want %d", player.CurrentHP, tt.wantHP)
			}
			if player.IsAlive() != tt.wantAlive {
				t.Errorf("ApplyDamageToPlayer() alive = %v, want %v", player.IsAlive(), tt.wantAlive)
			}
			if player.TempHP != tt.wantTempHP {
				t.Errorf("ApplyDamageToPlayer() temp HP = %d, want %d", player.TempHP, tt.wantTempHP)
			}
		})
	}
}

func TestCombatResolver_HealPlayer(t *testing.T) {
	resolver := NewCombatResolver()

	tests := []struct {
		name        string
		player      domain.PlayerState
		amount      int
		wantHP      int
	}{
		{
			name:   "normal heal",
			player: domain.PlayerState{CurrentHP: 5, MaxHP: 10},
			amount: 3,
			wantHP: 8,
		},
		{
			name:   "heal to full",
			player: domain.PlayerState{CurrentHP: 7, MaxHP: 10},
			amount: 5,
			wantHP: 10,
		},
		{
			name:   "heal from 0",
			player: domain.PlayerState{CurrentHP: 0, MaxHP: 10},
			amount: 3,
			wantHP: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := tt.player
			resolver.HealPlayer(&player, tt.amount)

			if player.CurrentHP != tt.wantHP {
				t.Errorf("HealPlayer() HP = %d, want %d", player.CurrentHP, tt.wantHP)
			}
		})
	}
}

func TestCombatResolver_SaveThrow(t *testing.T) {
	resolver := NewCombatResolver()

	tests := []struct {
		name     string
		ability  int      // ability modifier
		proficient bool
		profBonus int
		 DC       int
		roll     int
		wantSuccess bool
	}{
		{
			name:     "success",
			ability:  2,
			proficient: false,
			profBonus: 2,
			 DC:       12,
			roll:     10, // 10 + 2 = 12 >= 12 DC
			wantSuccess: true,
		},
		{
			name:     "success with proficiency",
			ability:  2,
			proficient: true,
			profBonus: 3,
			 DC:       15,
			roll:     10, // 10 + 2 + 3 = 15 >= 15 DC
			wantSuccess: true,
		},
		{
			name:     "failure",
			ability:  1,
			proficient: false,
			profBonus: 2,
			 DC:       15,
			roll:     10, // 10 + 1 = 11 < 15 DC
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolver.ResolveSaveThrow(tt.ability, tt.proficient, tt.profBonus, tt.DC, tt.roll)

			if result.Success != tt.wantSuccess {
				t.Errorf("ResolveSaveThrow() success = %v, want %v", result.Success, tt.wantSuccess)
			}
		})
	}
}

func TestCombatResolver_SkillCheck(t *testing.T) {
	resolver := NewCombatResolver()

	result := resolver.ResolveSkillCheck(3, true, 2, 15, 12)
	
	// 12 + 3 + 2 = 17 >= 15 DC
	if !result.Success {
		t.Error("SkillCheck should succeed: 12 + 3 + 2 = 17 >= 15")
	}
	if result.Total != 17 {
		t.Errorf("SkillCheck total = %d, want 17", result.Total)
	}

	result2 := resolver.ResolveSkillCheck(1, false, 2, 20, 10)
	
	// 10 + 1 = 11 < 20 DC
	if result2.Success {
		t.Error("SkillCheck should fail: 10 + 1 = 11 < 20")
	}
	if result2.Total != 11 {
		t.Errorf("SkillCheck total = %d, want 11", result2.Total)
	}
}
