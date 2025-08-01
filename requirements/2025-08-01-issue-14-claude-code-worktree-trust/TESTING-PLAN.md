# Testing Plan for Issue #14: Verify and Fix Automatic Claude Code Worktree Trust Configuration

## Overview

This testing plan covers the comprehensive verification and testing strategy for the automatic Claude Code worktree trust configuration feature. The goal is to ensure that worktree directories created by SBS are automatically trusted by Claude Code without requiring manual user intervention.

## Issue Summary

**GitHub Issue #14**: Verify and fix automatic Claude Code worktree trust configuration

**Current Implementation**: The `work-issue.sh` script includes a `update_claude_project_trust()` function that attempts to automatically configure trust settings in `~/.claude.json`.

**Problem Statement**: Need to verify that the automatic trust implementation works correctly and fix any issues preventing worktree directories from being automatically trusted by Claude Code.

## Current Trust Implementation Analysis

### Current Implementation Location
- **File**: `/home/jhjaggars/.work-issue-worktrees/work-issue/issue-14/work-issue.sh`
- **Function**: `update_claude_project_trust()` (lines 23-55)
- **Config Target**: `~/.claude.json`
- **Trust Method**: Sets `hasTrustDialogAccepted: true` and `hasCompletedProjectOnboarding: true`

### Current Implementation Details
```bash
# Function updates ~/.claude.json with:
{
  "projects": {
    "/path/to/worktree": {
      "allowedTools": [],
      "hasTrustDialogAccepted": true,
      "hasCompletedProjectOnboarding": true
    }
  }
}
```

## Test Structure and Organization

### 1. Trust Configuration Verification Tests

#### 1.1 Basic Trust Implementation Tests

**Test File**: `tests/trust/basic_trust_test.sh`

**Test Cases:**
```bash
test_trust_function_exists()
test_trust_function_requires_jq()
test_trust_function_creates_config_if_missing()
test_trust_function_updates_existing_config()
test_trust_function_handles_malformed_json()
test_trust_function_atomic_updates()
```

**Test Coverage:**
- Verify `update_claude_project_trust()` function exists and is callable
- Test dependency on `jq` command availability
- Test behavior when `~/.claude.json` doesn't exist
- Verify existing config files are properly updated
- Test handling of corrupted or malformed JSON configs
- Ensure atomic file updates (using temporary files)

#### 1.2 Config File Format Tests

**Test File**: `tests/trust/config_format_test.sh`

**Test Cases:**
```bash
test_config_file_location_correct()
test_config_structure_valid()
test_project_path_key_format()
test_trust_flags_set_correctly()
test_allowed_tools_array_present()
test_json_syntax_valid()
```

**Test Coverage:**
- Verify config file is created at correct location (`~/.claude.json`)
- Test JSON structure matches expected Claude Code format
- Verify project path is used as key correctly
- Confirm trust flags are set to `true`
- Test that `allowedTools` array is initialized
- Validate JSON syntax after modification

### 2. End-to-End Trust Workflow Tests

#### 2.1 Session Creation Trust Tests

**Test File**: `tests/e2e/session_trust_e2e_test.sh`

**Test Cases:**
```bash
test_new_session_creates_trust_config()
test_multiple_sessions_unique_trust_entries()
test_session_start_no_trust_prompt()
test_trust_persists_across_session_restarts()
test_trust_config_survives_sbs_updates()
```

**Test Coverage:**
- Test complete workflow: `sbs start N` → trust config created
- Verify multiple sessions create separate trust entries
- Confirm no trust dialog appears when starting Claude Code in worktree
- Test trust configuration persists after session stop/restart
- Verify trust config remains after SBS binary updates

#### 2.2 Claude Code Integration Tests

**Test File**: `tests/e2e/claude_integration_test.sh`

**Test Cases:**
```bash
test_claude_code_starts_without_trust_prompt()
test_claude_code_detects_trusted_directory()
test_claude_code_tool_access_granted()
test_claude_code_file_operations_allowed()
test_claude_code_git_operations_allowed()
```

**Test Coverage:**
- Launch Claude Code in new worktree, verify no trust prompt
- Test Claude Code recognizes directory as trusted
- Verify tool usage is allowed without additional prompts
- Test file read/write operations work without trust prompts
- Test git operations work without trust prompts

### 3. Claude Code Installation Scenarios

#### 3.1 Installation Method Tests

**Test File**: `tests/installation/claude_installation_test.sh`

**Test Cases:**
```bash
test_trust_with_npm_global_installation()
test_trust_with_yarn_global_installation()
test_trust_with_homebrew_installation()
test_trust_with_manual_binary_installation()
test_trust_with_dev_build_installation()
test_trust_with_multiple_installations()
```

**Test Coverage:**
- Test trust config works with npm global Claude Code installation
- Test trust config works with yarn global Claude Code installation
- Test trust config works with Homebrew Claude Code installation
- Test trust config works with manual binary installation
- Test trust config works with development builds
- Test behavior when multiple Claude Code installations exist

#### 3.2 Configuration Location Tests

**Test File**: `tests/installation/config_location_test.sh`

**Test Cases:**
```bash
test_home_claude_json_config()
test_xdg_config_claude_config()
test_app_specific_config_locations()
test_config_file_precedence()
test_config_migration_scenarios()
```

**Test Coverage:**
- Test `~/.claude.json` as primary config location
- Test XDG-compliant config directories
- Test application-specific config directories
- Verify config file precedence when multiple exist
- Test migration from old to new config formats

### 4. Edge Cases and Error Conditions

#### 4.1 Dependency and Environment Tests

**Test File**: `tests/edge_cases/dependency_test.sh`

**Test Cases:**
```bash
test_missing_jq_dependency()
test_jq_version_compatibility()
test_readonly_config_directory()
test_insufficient_disk_space()
test_concurrent_config_updates()
test_corrupted_config_recovery()
```

**Test Coverage:**
- Test graceful handling when `jq` is not installed
- Verify compatibility with different `jq` versions
- Test behavior when config directory is read-only
- Handle disk space exhaustion during config update
- Test concurrent updates from multiple SBS sessions
- Verify recovery from corrupted config files

#### 4.2 Path and Directory Tests

**Test File**: `tests/edge_cases/path_handling_test.sh`

**Test Cases:**
```bash
test_worktree_path_with_spaces()
test_worktree_path_with_special_chars()
test_deep_nested_worktree_paths()
test_symlinked_worktree_paths()
test_relative_vs_absolute_paths()
test_path_normalization()
```

**Test Coverage:**
- Test worktree paths containing spaces
- Test paths with special characters (quotes, ampersands, etc.)
- Test very deep directory structures
- Test symlinked worktree directories
- Verify relative paths are converted to absolute
- Test path normalization and canonicalization

### 5. Trust Configuration Research Tests

#### 5.1 Claude Code Config Format Research

**Test File**: `tests/research/claude_config_research.sh`

**Test Cases:**
```bash
test_discover_actual_config_location()
test_analyze_trust_dialog_behavior()
test_reverse_engineer_trust_mechanism()
test_document_config_schema()
test_validate_trust_implementation()
```

**Test Coverage:**
- Discover actual Claude Code configuration file locations
- Analyze when and how trust dialogs are triggered
- Reverse engineer the trust mechanism through testing
- Document the complete configuration schema
- Validate current implementation against actual behavior

#### 5.2 Alternative Trust Approaches

**Test File**: `tests/research/alternative_approaches_test.sh`

**Test Cases:**
```bash
test_command_line_trust_flags()
test_environment_variable_trust()
test_global_trust_configuration()
test_workspace_specific_trust()
test_programmatic_trust_api()
```

**Test Coverage:**
- Research command-line flags for trust configuration
- Test environment variables that might affect trust
- Investigate global trust settings
- Test workspace-specific trust mechanisms
- Research programmatic APIs for trust configuration

### 6. Regression and Compatibility Tests

#### 6.1 Backward Compatibility Tests

**Test File**: `tests/regression/compatibility_test.sh`

**Test Cases:**
```bash
test_existing_trust_configs_preserved()
test_manual_trust_settings_respected()
test_mixed_trust_configuration_handling()
test_config_format_version_compatibility()
test_claude_code_version_compatibility()
```

**Test Coverage:**
- Verify existing trust configurations are not overwritten
- Test that manually set trust settings are preserved
- Handle mixed automatic and manual trust configurations
- Test compatibility across Claude Code config format versions
- Verify compatibility across Claude Code application versions

#### 6.2 Performance and Resource Tests

**Test File**: `tests/regression/performance_test.sh`

**Test Cases:**
```bash
test_trust_update_performance()
test_config_file_size_growth()
test_memory_usage_impact()
test_concurrent_session_handling()
test_large_project_count_handling()
```

**Test Coverage:**
- Measure time taken for trust configuration updates
- Monitor config file size growth with many projects
- Test memory usage impact of trust configuration
- Verify performance with many concurrent sessions
- Test handling of configs with large numbers of projects

### 7. Test Implementation Requirements

#### 7.1 Test Environment Setup

**Setup Script**: `tests/setup/test_environment_setup.sh`

**Requirements:**
```bash
# Test environment isolation
create_isolated_test_environment()
backup_existing_claude_config()
setup_mock_claude_installation()
prepare_test_worktree_locations()
install_test_dependencies()
```

**Environment Setup:**
- Isolated test environment with temporary home directory
- Backup and restore existing Claude Code configurations
- Mock Claude Code installations for testing
- Predefined test worktree locations and structures
- Installation of required test dependencies (jq, etc.)

#### 7.2 Test Data and Fixtures

**Test Data Directory**: `tests/fixtures/`

**Fixture Files:**
```
tests/fixtures/
├── claude-configs/
│   ├── empty-config.json
│   ├── basic-config.json
│   ├── complex-config.json
│   ├── malformed-config.json
│   └── legacy-format-config.json
├── worktree-paths/
│   ├── simple-paths.txt
│   ├── paths-with-spaces.txt
│   ├── paths-with-special-chars.txt
│   └── deep-nested-paths.txt
└── expected-outputs/
    ├── trust-config-after-update.json
    ├── multiple-projects-config.json
    └── error-messages.txt
```

#### 7.3 Mock and Stub Requirements

**Mock Files**: `tests/mocks/`

**Mock Components:**
```bash
# Mock Claude Code binary
mock_claude_code_binary()
mock_trust_dialog_behavior()
mock_config_file_operations()
mock_jq_command_behavior()
mock_file_system_operations()
```

### 8. Test Execution Strategy

#### 8.1 Test Categories and Execution Order

**Phase 1: Unit Tests**
1. Trust function unit tests
2. Config format validation tests
3. Error handling tests

**Phase 2: Integration Tests**
4. Claude Code integration tests
5. Installation scenario tests
6. End-to-end trust workflow tests

**Phase 3: Research and Discovery**
7. Configuration research tests
8. Alternative approach investigation
9. Compatibility analysis

**Phase 4: Regression and Performance**
10. Backward compatibility tests
11. Performance and resource tests
12. Load testing with multiple sessions

#### 8.2 Test Execution Commands

```bash
# Run all trust-related tests
./tests/run_trust_tests.sh

# Run specific test categories
./tests/run_trust_tests.sh --category unit
./tests/run_trust_tests.sh --category integration
./tests/run_trust_tests.sh --category research
./tests/run_trust_tests.sh --category regression

# Run tests for specific Claude Code installation
./tests/run_trust_tests.sh --claude-install npm
./tests/run_trust_tests.sh --claude-install homebrew

# Run tests with different worktree scenarios
./tests/run_trust_tests.sh --scenario single-session
./tests/run_trust_tests.sh --scenario multiple-sessions
./tests/run_trust_tests.sh --scenario concurrent-sessions

# Generate test reports
./tests/run_trust_tests.sh --report html
./tests/run_trust_tests.sh --report json
```

### 9. Validation Criteria

#### 9.1 Success Criteria for Each Test Category

**Trust Configuration Tests:**
- Trust function executes without errors
- Config file is created/updated correctly
- JSON structure matches expected format
- Project path is correctly stored as key

**End-to-End Tests:**
- New sessions automatically create trust configuration
- Claude Code starts without trust prompts in worktrees
- Trust configuration persists across session lifecycles
- Multiple sessions create independent trust entries

**Installation Compatibility:**
- Trust configuration works with all major Claude Code installation methods
- Configuration file is found and updated regardless of installation type
- Trust mechanism works consistently across installations

**Edge Cases:**
- Graceful handling of missing dependencies
- Proper error messages for configuration failures
- Recovery from corrupted configuration files
- Handling of unusual path formats and characters

#### 9.2 Failure Criteria and Root Cause Analysis

**Critical Failures:**
- Trust prompts still appear after automatic configuration
- Configuration file corruption or data loss
- Complete failure to update configuration
- Breaking existing Claude Code functionality

**Root Cause Analysis Process:**
1. Identify specific failure point in trust workflow
2. Analyze Claude Code configuration mechanism
3. Compare expected vs. actual configuration format
4. Investigate Claude Code version-specific differences
5. Document findings and propose solutions

### 10. Test Reporting and Documentation

#### 10.1 Test Report Structure

**Report File**: `tests/reports/trust-configuration-test-report.md`

**Report Sections:**
```markdown
# Claude Code Trust Configuration Test Report

## Executive Summary
- Overall test status
- Critical issues identified
- Recommendations

## Test Results by Category
- Unit test results
- Integration test results
- Research findings
- Performance metrics

## Configuration Analysis
- Current implementation assessment
- Claude Code behavior analysis
- Recommended configuration format

## Issues and Recommendations
- Critical issues requiring immediate attention
- Enhancement recommendations
- Implementation alternatives

## Appendices
- Detailed test logs
- Configuration file examples
- Error message catalog
```

#### 10.2 Issue Documentation Requirements

**For Working Implementation:**
- Document verified trust configuration mechanism
- Provide examples of successful configurations
- Create troubleshooting guide for common issues

**For Broken Implementation:**
- Root cause analysis of failure points
- Detailed explanation of why current approach fails
- Research findings on correct Claude Code trust mechanism
- Recommended implementation approach
- Limitations and constraints documentation

### 11. Test Automation and CI Integration

#### 11.1 Automated Test Execution

**GitHub Actions Workflow**: `.github/workflows/trust-configuration-tests.yml`

```yaml
name: Claude Code Trust Configuration Tests

on:
  push:
    paths:
      - 'work-issue.sh'
      - 'tests/trust/**'
  pull_request:
    paths:
      - 'work-issue.sh'
      - 'tests/trust/**'

jobs:
  trust-tests:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        claude-install-method: [npm, manual, mock]
        test-scenario: [single-session, multiple-sessions]
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v3
      
      - name: Setup test environment
        run: ./tests/setup/test_environment_setup.sh
      
      - name: Install Claude Code (${{ matrix.claude-install-method }})
        run: ./tests/setup/install_claude_${{ matrix.claude-install-method }}.sh
      
      - name: Run trust configuration tests
        run: |
          ./tests/run_trust_tests.sh \
            --claude-install ${{ matrix.claude-install-method }} \
            --scenario ${{ matrix.test-scenario }}
      
      - name: Generate test report
        run: ./tests/run_trust_tests.sh --report json
      
      - name: Upload test results
        uses: actions/upload-artifact@v3
        with:
          name: trust-test-results-${{ matrix.claude-install-method }}-${{ matrix.test-scenario }}
          path: tests/reports/
```

#### 11.2 Local Development Testing

**Development Scripts:**
```bash
# Quick trust test during development
./tests/quick_trust_test.sh

# Test specific trust function changes
./tests/test_trust_function.sh

# Verify trust config after SBS changes
./tests/verify_trust_integration.sh

# Reset test environment
./tests/reset_test_environment.sh
```

### 12. Dependencies and Prerequisites

#### 12.1 Test Dependencies

**Required Tools:**
- `jq` (JSON processing)
- `bash` (version 4.0+)
- `curl` or `wget` (for Claude Code installation testing)
- `npm` or `yarn` (for npm/yarn installation testing)
- `git` (for worktree operations)
- `tmux` (for session testing)

**Optional Tools:**
- `homebrew` (for Homebrew installation testing)
- `docker` (for isolated testing environments)
- `shellcheck` (for shell script validation)

#### 12.2 Test Environment Requirements

**System Requirements:**
- Linux or macOS operating system
- Bash shell (version 4.0 or later)
- At least 1GB free disk space for test worktrees
- Network access for Claude Code installation testing

**Permissions Requirements:**
- Write access to home directory for config file testing
- Ability to create temporary directories
- Permission to install software (for installation testing)

## Implementation Roadmap

### Phase 1: Test Infrastructure (Week 1)
1. Create test directory structure
2. Implement test environment setup scripts
3. Create basic test fixtures and mocks
4. Set up automated test execution framework

### Phase 2: Core Trust Testing (Week 2)
1. Implement unit tests for trust function
2. Create config format validation tests
3. Build end-to-end trust workflow tests
4. Test multiple worktree scenarios

### Phase 3: Installation and Integration (Week 3)
1. Test different Claude Code installation methods
2. Verify trust configuration across installations
3. Research actual Claude Code trust mechanism
4. Document configuration format requirements

### Phase 4: Edge Cases and Research (Week 4)
1. Implement edge case and error condition tests
2. Conduct deep research into Claude Code trust behavior
3. Investigate alternative trust configuration approaches
4. Performance and compatibility testing

### Phase 5: Analysis and Recommendations (Week 5)
1. Analyze all test results
2. Perform root cause analysis of any failures
3. Document working solutions or limitations
4. Create comprehensive test report and recommendations

## Success Metrics

### Quantitative Metrics
- **Test Coverage**: 95%+ of trust-related code paths covered
- **Success Rate**: 100% of trust configurations should work without user intervention
- **Performance**: Trust configuration should complete in <100ms
- **Compatibility**: 100% compatibility across supported Claude Code installation methods

### Qualitative Metrics
- **User Experience**: No manual trust prompts for SBS-created worktrees
- **Reliability**: Trust configuration works consistently across environments
- **Maintainability**: Trust mechanism is well-documented and testable
- **Robustness**: Graceful handling of edge cases and error conditions

This comprehensive testing plan ensures thorough verification of the Claude Code worktree trust configuration feature, covering all aspects from basic functionality to complex edge cases and integration scenarios. The phased approach allows for systematic testing and iterative improvement of the trust configuration mechanism.