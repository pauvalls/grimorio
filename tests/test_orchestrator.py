#!/usr/bin/env python3
"""Tests for grimorio-architect.md — verifies end-to-end orchestration and progress reporting."""

import re
import sys
from pathlib import Path

ARCHITECT_PATH = Path(__file__).parent.parent / "agents" / "grimorio-architect.md"

def test_file_exists():
    """Test that architect file exists."""
    assert ARCHITECT_PATH.exists(), f"File not found: {ARCHITECT_PATH}"
    print("✅ File exists")

def test_has_report_progress_responsibility():
    """Test that architect has reporting progress as a responsibility."""
    content = ARCHITECT_PATH.read_text()
    assert "REPORT PROGRESS" in content, "Missing REPORT PROGRESS in responsibilities"
    assert "report progress to the user after each phase" in content.lower()
    print("✅ Has report progress responsibility")

def test_has_phase_4b_report():
    """Test that Phase 4b (content status report) exists."""
    content = ARCHITECT_PATH.read_text()
    assert "Phase 4b: Report Content Status to User" in content, "Missing Phase 4b"
    assert "Fase 3 Completada" in content, "Missing Spanish report header"
    assert "delegation_read" in content, "Missing delegation_read instruction"
    print("✅ Has Phase 4b content report")

def test_has_phase_6b_report():
    """Test that Phase 6b (acts status report) exists."""
    content = ARCHITECT_PATH.read_text()
    assert "Phase 6b: Report Acts Status to User" in content, "Missing Phase 6b"
    assert "Fase 5 Completada" in content, "Missing Spanish report header"
    print("✅ Has Phase 6b acts report")

def test_has_phase_8b_report():
    """Test that Phase 8b (SVGs + artist status report) exists."""
    content = ARCHITECT_PATH.read_text()
    assert "Phase 8b: Report SVGs + Artist Status to User" in content, "Missing Phase 8b"
    print("✅ Has Phase 8b SVGs report")

def test_has_phase_9_progress_reports():
    """Test that Phase 9 has start and end progress reports."""
    content = ARCHITECT_PATH.read_text()
    assert "Fase 9 Iniciada" in content, "Missing Phase 9 start report"
    assert "Fase 9 Completada" in content, "Missing Phase 9 completion report"
    print("✅ Has Phase 9 progress reports")

def test_has_phase_11b_report():
    """Test that Phase 11b (references status report) exists."""
    content = ARCHITECT_PATH.read_text()
    assert "Phase 11b: Report References Status to User" in content, "Missing Phase 11b"
    assert "Fase 10 Completada" in content, "Missing Spanish report header"
    print("✅ Has Phase 11b references report")

def test_has_phase_12_progress_reports():
    """Test that Phase 12 has start and end progress reports."""
    content = ARCHITECT_PATH.read_text()
    assert "Fase 12" in content, "Missing Phase 12 report"
    assert "PDF Compilado Exitosamente" in content, "Missing PDF completion report"
    print("✅ Has Phase 12 progress reports")

def test_has_final_report():
    """Test that final user-visible report exists."""
    content = ARCHITECT_PATH.read_text()
    assert "Phase 13: Final Report to User" in content, "Missing Phase 13"
    assert 'Campaña' in content, "Missing final Spanish report"
    print("✅ Has final report to user")

def test_phases_in_correct_order():
    """Test that phases appear in correct sequential order."""
    content = ARCHITECT_PATH.read_text()
    
    phase_pattern = r'### Phase (\d+)\w*:'
    phases = [(int(m.group(1)), m.start()) for m in re.finditer(phase_pattern, content)]
    
    positions = [pos for _, pos in phases]
    assert positions == sorted(positions), "Phases are not in order"
    
    phase_numbers = sorted(set(num for num, _ in phases))
    expected = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13]
    assert phase_numbers == expected, f"Missing phases. Expected {expected}, got {phase_numbers}"
    print("✅ Phases in correct order")

def test_spanish_reports():
    """Test that reports use Spanish language."""
    content = ARCHITECT_PATH.read_text()
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
    content = ARCHITECT_PATH.read_text()
    count = content.count("delegation_read")
    assert count >= 5, f"delegation_read should appear at least 5 times, found {count}"
    print(f"✅ delegation_read used {count} times")

def test_no_ask_user_questions_after_phase1():
    """Test that architect does not ask user questions after Phase 1."""
    content = ARCHITECT_PATH.read_text()
    assert "NEVER ask the user questions after Phase 1" in content, "Missing rule"
    print("✅ Rule: never ask user questions after Phase 1")

def test_uses_general_agent_for_content():
    """Test that content generation uses general agent type."""
    content = ARCHITECT_PATH.read_text()
    count = content.count('delegate(agent="general"')
    assert count >= 4, f"Should delegate to general agent at least 4 times, found {count}"
    print(f"✅ Uses general agent for content ({count} times)")

def test_has_mcp_tools():
    """Test that architect has MCP tools configured."""
    content = ARCHITECT_PATH.read_text()
    assert "grimorio_generate_image" in content, "Missing generate_image tool"
    assert "grimorio_generate_map" in content, "Missing generate_map tool"
    assert "grimorio_generate_divider" in content, "Missing generate_divider tool"
    assert "grimorio_compile_pdf" in content, "Missing compile_pdf tool"
    print("✅ Has all required MCP tools")

def run_all_tests():
    """Run all tests."""
    tests = [
        test_file_exists,
        test_has_report_progress_responsibility,
        test_has_phase_4b_report,
        test_has_phase_6b_report,
        test_has_phase_8b_report,
        test_has_phase_9_progress_reports,
        test_has_phase_11b_report,
        test_has_phase_12_progress_reports,
        test_has_final_report,
        test_phases_in_correct_order,
        test_spanish_reports,
        test_delegation_read_usage,
        test_no_ask_user_questions_after_phase1,
        test_uses_general_agent_for_content,
        test_has_mcp_tools,
    ]
    
    passed = 0
    failed = 0
    
    print("\n🧪 Running grimorio-architect orchestration tests...\n")
    
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
