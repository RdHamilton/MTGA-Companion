#!/usr/bin/env python3
"""Script to add issues to projects, set dates, and assign phases/milestones."""

import subprocess
import json
import sys
from datetime import datetime

# Milestone mapping
milestone_map = {
    "Security Fix": 2, "Database Improvements": 3, "Poller Enhancements": 4,
    "Card Metadata Source": 5, "Full Integration": 6,
    "Basic Statistics": 7, "Advanced Analytics": 8, "Rank Features": 9,
    "Draft Storage": 10, "Draft Display": 11, "Draft Analytics": 12,
    "Collection Features": 13, "Deck Analysis": 14, "Collection-Deck Integration": 15,
    "GUI Foundation": 16, "Statistics GUI": 17, "Charts & Graphs": 18, "Advanced GUI Features": 19,
    "Export Features": 20, "External Integration": 21
}

# Project mapping
project_map = {
    "Security & Infrastructure": {"number": 2, "id": "PVT_kwHOABsZ684BHe6N"},
    "Card Metadata Integration": {"number": 3, "id": "PVT_kwHOABsZ684BHe6O"},
    "Statistics Enhancements": {"number": 4, "id": "PVT_kwHOABsZ684BHe6P"},
    "Draft Features": {"number": 5, "id": "PVT_kwHOABsZ684BHe6Q"},
    "Collection & Deck Management": {"number": 6, "id": "PVT_kwHOABsZ684BHe6S"},
    "Fyne GUI Foundation": {"number": 7, "id": "PVT_kwHOABsZ684BHe6V"},
    "GUI Features": {"number": 8, "id": "PVT_kwHOABsZ684BHe6W"},
    "Export & Integration": {"number": 9, "id": "PVT_kwHOABsZ684BHe6X"}
}

# Issue assignments: (issue_num, phase, milestone, date)
assignments = {
    2: [(31, "Phase 1: Security", "Security Fix", "2025-11-07"),
        (59, "Phase 2: Database", "Database Improvements", "2025-11-08"),
        (60, "Phase 2: Database", "Database Improvements", "2025-11-09"),
        (65, "Phase 3: Poller", "Poller Enhancements", "2025-11-10"),
        (66, "Phase 3: Poller", "Poller Enhancements", "2025-11-11"),
        (67, "Phase 3: Poller", "Poller Enhancements", "2025-11-12"),
        (68, "Phase 3: Poller", "Poller Enhancements", "2025-11-13"),
        (69, "Phase 3: Poller", "Poller Enhancements", "2025-11-14")],
    3: [(71, "Phase 1: Foundation", "Card Metadata Source", "2025-11-15"),
        (79, "Phase 2: Integration", "Full Integration", "2025-11-16"),
        (118, "Phase 2: Integration", "Full Integration", "2025-11-17")],
    4: [(38, "Phase 1: Core Statistics", "Basic Statistics", "2025-11-18"),
        (39, "Phase 1: Core Statistics", "Basic Statistics", "2025-11-19"),
        (40, "Phase 1: Core Statistics", "Basic Statistics", "2025-11-20"),
        (41, "Phase 1: Core Statistics", "Basic Statistics", "2025-11-21"),
        (42, "Phase 1: Core Statistics", "Basic Statistics", "2025-11-22"),
        (43, "Phase 1: Core Statistics", "Basic Statistics", "2025-11-23"),
        (44, "Phase 1: Core Statistics", "Basic Statistics", "2025-11-24"),
        (45, "Phase 1: Core Statistics", "Basic Statistics", "2025-11-25"),
        (46, "Phase 1: Core Statistics", "Basic Statistics", "2025-11-26"),
        (47, "Phase 2: Advanced Analytics", "Advanced Analytics", "2025-11-27"),
        (48, "Phase 1: Core Statistics", "Basic Statistics", "2025-11-28"),
        (49, "Phase 1: Core Statistics", "Basic Statistics", "2025-11-29"),
        (57, "Phase 2: Advanced Analytics", "Advanced Analytics", "2025-11-30"),
        (72, "Phase 1: Core Statistics", "Basic Statistics", "2025-12-01"),
        (76, "Phase 1: Core Statistics", "Basic Statistics", "2025-12-02"),
        (81, "Phase 1: Core Statistics", "Basic Statistics", "2025-12-03"),
        (87, "Phase 2: Advanced Analytics", "Advanced Analytics", "2025-12-04"),
        (88, "Phase 2: Advanced Analytics", "Advanced Analytics", "2025-12-05"),
        (89, "Phase 2: Advanced Analytics", "Advanced Analytics", "2025-12-06"),
        (90, "Phase 2: Advanced Analytics", "Advanced Analytics", "2025-12-07"),
        (91, "Phase 2: Advanced Analytics", "Advanced Analytics", "2025-12-08"),
        (92, "Phase 2: Advanced Analytics", "Advanced Analytics", "2025-12-09"),
        (94, "Phase 2: Advanced Analytics", "Advanced Analytics", "2025-12-10"),
        (95, "Phase 2: Advanced Analytics", "Advanced Analytics", "2025-12-11"),
        (96, "Phase 2: Advanced Analytics", "Advanced Analytics", "2025-12-12"),
        (97, "Phase 2: Advanced Analytics", "Advanced Analytics", "2025-12-13"),
        (98, "Phase 2: Advanced Analytics", "Advanced Analytics", "2025-12-14"),
        (99, "Phase 2: Advanced Analytics", "Advanced Analytics", "2025-12-15"),
        (100, "Phase 1: Core Statistics", "Basic Statistics", "2025-12-16"),
        (102, "Phase 3: Rank Tracking", "Rank Features", "2025-12-17"),
        (103, "Phase 3: Rank Tracking", "Rank Features", "2025-12-18"),
        (104, "Phase 3: Rank Tracking", "Rank Features", "2025-12-19"),
        (105, "Phase 3: Rank Tracking", "Rank Features", "2025-12-20"),
        (106, "Phase 3: Rank Tracking", "Rank Features", "2025-12-21"),
        (107, "Phase 3: Rank Tracking", "Rank Features", "2025-12-22"),
        (108, "Phase 3: Rank Tracking", "Rank Features", "2025-12-23"),
        (109, "Phase 3: Rank Tracking", "Rank Features", "2025-12-24"),
        (110, "Phase 3: Rank Tracking", "Rank Features", "2025-12-25")],
    5: [(112, "Phase 1: Storage", "Draft Storage", "2025-12-26"),
        (113, "Phase 2: Display", "Draft Display", "2025-12-27"),
        (114, "Phase 2: Display", "Draft Display", "2025-12-28"),
        (115, "Phase 3: Analytics", "Draft Analytics", "2025-12-29"),
        (116, "Phase 2: Display", "Draft Display", "2025-12-30"),
        (117, "Phase 2: Display", "Draft Display", "2025-12-31"),
        (118, "Phase 2: Display", "Draft Display", "2026-01-01"),
        (119, "Phase 3: Analytics", "Draft Analytics", "2026-01-02"),
        (120, "Phase 3: Analytics", "Draft Analytics", "2026-01-03"),
        (121, "Phase 3: Analytics", "Draft Analytics", "2026-01-04"),
        (122, "Phase 3: Analytics", "Draft Analytics", "2026-01-05")],
    6: [(73, "Phase 1: Collection", "Collection Features", "2026-01-06"),
        (74, "Phase 3: Integration", "Collection-Deck Integration", "2026-01-07"),
        (75, "Phase 3: Integration", "Collection-Deck Integration", "2026-01-08"),
        (76, "Phase 1: Collection", "Collection Features", "2026-01-09"),
        (77, "Phase 1: Collection", "Collection Features", "2026-01-10"),
        (80, "Phase 2: Deck Analysis", "Deck Analysis", "2026-01-11"),
        (82, "Phase 2: Deck Analysis", "Deck Analysis", "2026-01-12"),
        (83, "Phase 3: Integration", "Collection-Deck Integration", "2026-01-13"),
        (84, "Phase 3: Integration", "Collection-Deck Integration", "2026-01-14"),
        (85, "Phase 3: Integration", "Collection-Deck Integration", "2026-01-15")],
    7: [(50, "Phase 1: Foundation", "GUI Foundation", "2026-01-16"),
        (53, "Phase 1: Foundation", "GUI Foundation", "2026-01-17")],
    8: [(51, "Phase 1: Statistics GUI", "Statistics GUI", "2026-01-18"),
        (52, "Phase 2: Visualizations", "Charts & Graphs", "2026-01-19"),
        (54, "Phase 2: Visualizations", "Charts & Graphs", "2026-01-20"),
        (55, "Phase 1: Statistics GUI", "Statistics GUI", "2026-01-21"),
        (56, "Phase 3: Advanced Features", "Advanced GUI Features", "2026-01-22"),
        (61, "Phase 3: Advanced Features", "Advanced GUI Features", "2026-01-23"),
        (62, "Phase 3: Advanced Features", "Advanced GUI Features", "2026-01-24"),
        (63, "Phase 3: Advanced Features", "Advanced GUI Features", "2026-01-25"),
        (77, "Phase 1: Statistics GUI", "Statistics GUI", "2026-01-26"),
        (90, "Phase 2: Visualizations", "Charts & Graphs", "2026-01-27"),
        (98, "Phase 2: Visualizations", "Charts & Graphs", "2026-01-28"),
        (105, "Phase 1: Statistics GUI", "Statistics GUI", "2026-01-29"),
        (110, "Phase 2: Visualizations", "Charts & Graphs", "2026-01-30"),
        (113, "Phase 1: Statistics GUI", "Statistics GUI", "2026-01-31"),
        (114, "Phase 1: Statistics GUI", "Statistics GUI", "2026-02-01"),
        (117, "Phase 2: Visualizations", "Charts & Graphs", "2026-02-02")],
    9: [(42, "Phase 1: Export", "Export Features", "2026-02-03"),
        (58, "Phase 2: Integration", "External Integration", "2026-02-04"),
        (73, "Phase 1: Export", "Export Features", "2026-02-05"),
        (82, "Phase 1: Export", "Export Features", "2026-02-06"),
        (92, "Phase 1: Export", "Export Features", "2026-02-07"),
        (97, "Phase 1: Export", "Export Features", "2026-02-08"),
        (116, "Phase 1: Export", "Export Features", "2026-02-09")]
}

def get_issue_node_id(issue_num):
    """Get the GraphQL node ID for an issue."""
    query = f'''
    {{
      repository(owner: "RdHamilton", name: "MTGA-Companion") {{
        issue(number: {issue_num}) {{
          id
        }}
      }}
    }}
    '''
    result = subprocess.run(
        ['gh', 'api', 'graphql', '-f', f'query={query}'],
        capture_output=True, text=True, check=True
    )
    data = json.loads(result.stdout)
    return data['data']['repository']['issue']['id']

def add_issue_to_project(project_id, issue_node_id):
    """Add an issue to a project."""
    mutation = f'''
    mutation {{
      addProjectV2ItemById(input: {{
        projectId: "{project_id}"
        contentId: "{issue_node_id}"
      }}) {{
        item {{
          id
        }}
      }}
    }}
    '''
    try:
        result = subprocess.run(
            ['gh', 'api', 'graphql', '-f', f'query={mutation}'],
            capture_output=True, text=True, check=True
        )
        data = json.loads(result.stdout)
        return data['data']['addProjectV2ItemById']['item']['id']
    except subprocess.CalledProcessError as e:
        # Item might already be in project
        if "already exists" in e.stderr.lower() or "already added" in e.stderr.lower():
            return None
        raise

def set_issue_milestone(issue_num, milestone_num):
    """Set milestone on an issue."""
    try:
        subprocess.run(
            ['gh', 'issue', 'edit', str(issue_num), '--milestone', str(milestone_num)],
            capture_output=True, text=True, check=True
        )
        return True
    except subprocess.CalledProcessError:
        return False

def set_issue_due_date(issue_num, date_str):
    """Set due date on an issue (using issue body or labels)."""
    # GitHub issues don't have built-in due dates, so we'll add it to the issue body
    try:
        # Get current issue body
        result = subprocess.run(
            ['gh', 'issue', 'view', str(issue_num), '--json', 'body'],
            capture_output=True, text=True, check=True
        )
        data = json.loads(result.stdout)
        body = data.get('body', '') or ''
        
        # Add due date if not present
        if f"**Due Date:** {date_str}" not in body:
            if body:
                body = f"{body}\n\n**Due Date:** {date_str}"
            else:
                body = f"**Due Date:** {date_str}"
            
            subprocess.run(
                ['gh', 'issue', 'edit', str(issue_num), '--body', body],
                capture_output=True, text=True, check=True
            )
        return True
    except subprocess.CalledProcessError:
        return False

def main():
    total = sum(len(issues) for issues in assignments.values())
    processed = 0
    errors = []
    
    for project_num, issues in assignments.items():
        project_name = [k for k, v in project_map.items() if v["number"] == project_num][0]
        project_id = project_map[project_name]["id"]
        
        print(f"\nProcessing {project_name} ({len(issues)} issues)...")
        
        for issue_num, phase, milestone, date_str in issues:
            try:
                print(f"  Issue #{issue_num}...", end=" ", flush=True)
                
                # Get issue node ID
                issue_node_id = get_issue_node_id(issue_num)
                
                # Add to project
                item_id = add_issue_to_project(project_id, issue_node_id)
                if item_id:
                    print("added to project", end=", ", flush=True)
                else:
                    print("already in project", end=", ", flush=True)
                
                # Set milestone
                if set_issue_milestone(issue_num, milestone_map[milestone]):
                    print("milestone set", end=", ", flush=True)
                
                # Set due date
                if set_issue_due_date(issue_num, date_str):
                    print("due date set", end="", flush=True)
                
                print(" ✓")
                processed += 1
                
            except Exception as e:
                print(f" ✗ Error: {e}")
                errors.append((issue_num, str(e)))
    
    print(f"\n\nCompleted: {processed}/{total} issues processed")
    if errors:
        print(f"\nErrors ({len(errors)}):")
        for issue_num, error in errors:
            print(f"  Issue #{issue_num}: {error}")

if __name__ == "__main__":
    main()


