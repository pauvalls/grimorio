#!/usr/bin/env python3
"""Tests for grimorio-orchestrator.md structure and progress reporting requirements."""

import re
import sys
from pathlib import Path

ORCHESTRATOR_PATH = Path(__file__).parent.parent / "agents" / "grimorio-orchestrator.md"

def test_file_exists():
    """Test that orchestrator file exists."""
    assert ORCHESTRATOR_PATH.exists(), f"File not found: {ORCHESTRATOR_PATH}"
    print("✅ File exists")

def test_has_report_progress_responsibility():
    """Test that orchestrator has reporting progress as a responsibility."""
    content = ORCHESTRATOR_PATH.read_text()
    assert "REPORT PROGRESS" in content, "Missing REPORT PROGRESS in responsibilities"
    assert "REPORT PROGRESS to the user after each phase completes" in content
    print("✅ Has report progress responsibility")

def test_has_phase_2b_report():
    """Test that Phase 2b (content status report) exists."""
    content = ORCHESTRATOR_PATH.read_text()
    assert "Phase 2b: Report Content Status to User" in content, "Missing Phase 2b"
    assert "Fase 1 Completada" in content, "Missing Spanish report header for Phase 1"
    assert "delegation_read" in content, "Missing delegation_read instruction"
    print("✅ Has Phase 2b content report")

def test_has_phase_4b_report():
    """Test that Phase 4b (acts status report) exists."""
    content = ORCHESTRATOR_PATH.read_text()
    assert "Phase 4b: Report Acts Status to User" in content, "Missing Phase 4b"
    assert "Fase 3 Completada" in content, "Missing Spanish report header for Phase 3"
    print("✅ Has Phase 4b acts report")

def test_has_phase_6b_report():
    """Test that Phase 6b (SVGs + artist status report) exists."""
    content = ORCHESTRATOR_PATH.read_text()
    assert "Phase 6b: Report SVGs + Artist Status to User" in content, "Missing Phase 6b"
    assert "Fase 5 Completada" in content, "Missing Spanish report header for Phase 5"
    print("✅ Has Phase 6b SVGs report")

def test_has_phase_7_progress_reports():
    """Test that Phase 7 has start and end progress reports."""
    content = ORCHESTRATOR_PATH.read_text()
    assert "Fase 7 Iniciada" in content, "Missing Phase 7 start report"
    assert "Fase 7 Completada" in content, "Missing Phase 7 completion report"
    print("✅ Has Phase 7 progress reports")

def test_has_phase_9b_report():
    """Test that Phase 9b (references status report) exists."""
    content = ORCHESTRATOR_PATH.read_text()
    assert "Phase 9b: Report References Status to User" in content, "Missing Phase 9b"
    assert "Fase 8 Completada" in content, "Missing Spanish report header for Phase 8"
    print("✅ Has Phase 9b references report")

def test_has_phase_10_progress_reports():
    """Test that Phase 10 has start and end progress reports."""
    content = ORCHESTRATOR_PATH.read_text()
    assert "Fase 10" in content, "Missing Phase 10 report"
    assert "PDF Compilado Exitosamente" in content, "Missing PDF completion report"
    print("✅ Has Phase 10 progress reports")

def test_has_final_report():
    """Test that final user-visible report exists."""
    content = ORCHESTRATOR_PATH.read_text()
    assert "Phase 11: Final Report to Parent AND User" in content, "Missing Phase 11"
    assert 'Campaña "{campaign_name}" Completada' in content, "Missing final Spanish report"
    print("✅ Has final report to user")

def test_phases_in_correct_order():
    """Test that phases appear in correct sequential order."""
    content = ORCHESTRATOR_PATH.read_text()
    
    phase_pattern = r'### Phase (\d+)\w*:'
    phases = [(int(m.group(1)), m.start()) for m in re.finditer(phase_pattern, content)]
    
    # Check that phases are in ascending order by position
    positions = [pos for _, pos in phases]
    assert positions == sorted(positions), "Phases are not in order"
    
    # Check that we have all main phases (1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11)
    phase_numbers = sorted(set(num for num, _ in phases))
    expected = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11]
    assert phase_numbers == expected, f"Missing phases. Expected {expected}, got {phase_numbers}"
    print("✅ Phases in correct order")

def test_spanish_reports():
    """Test that reports use Spanish language."""
    content = ORCHESTRATOR_PATH.read_text()
    spanish_markers = [
        "Fase",
        "Completada",
        "Iniciando",
        "Generación",
        "Campaña",
        "Contenido Generado",
        "Estado"
    ]
    for marker in spanish_markers:
        assert marker in content, f"Missing Spanish marker: {marker}"
    print("✅ Reports use Spanish language")

def test_delegation_read_usage():
    """Test that delegation_read is used in monitoring phases."""
    content = ORCHESTRATOR_PATH.read_text()
    # Should appear multiple times (in monitoring + report phases)
    count = content.count("delegation_read")
    assert count >= 5, f"delegation_read should appear at least 5 times, found {count}"
    print(f"✅ delegation_read used {count} times")

def test_no_ask_user_questions():
    """Test that orchestrator does not ask user questions."""
    content = ORCHESTRATOR_PATH.read_text()
    assert "NEVER ask the user questions" in content, "Missing 'never ask questions' rule"
    print("✅ Rule: never ask user questions")

def run_all_tests():
    """Run all tests."""
    tests = [
        test_file_exists,
        test_has_report_progress_responsibility,
        test_has_phase_2b_report,
        test_has_phase_4b_report,
        test_has_phase_6b_report,
        test_has_phase_7_progress_reports,
        test_has_phase_9b_report,
        test_has_phase_10_progress_reports,
        test_has_final_report,
        test_phases_in_correct_order,
        test_spanish_reports,
        test_delegation_read_usage,
        test_no_ask_user_questions,
    ]
    
    passed = 0
    failed = 0
    
    print("\n🧪 Running grimorio-orchestrator tests...\n")
    
    for test in tests:
        try:
            test()
            passed += 1
        except AssertionError as e:
            print(f"❌ {test.__name__}: {e}")
            failed += 1
        except Exception as e:
            print(f"💥 {test.__name__}: {e}")
            failed += 1
    
    print(f"\n{'='*50}")
    print(f"Results: {passed} passed, {failed} failed")
    
    if failed > 0:
        sys.exit(1)
    else:
        print("\n🎉 All tests passed!")

if __name__ == "__main__":
    run_all_tests()
