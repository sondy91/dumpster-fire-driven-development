# config.py - Centralized configuration
from dataclasses import dataclass


@dataclass
class Contributor:
    name: str
    commitment_percentage: float
    pto_days: int = 0


@dataclass
class SprintConfig:
    length_days: int
    holidays_in_sprint: int
    contributors: list[Contributor]


# --- CURRENT SPRINT CONFIGURATION ---
# ⬇️ EDIT THESE VALUES BEFORE EACH SPRINT ⬇️
CURRENT_SPRINT = SprintConfig(
    length_days=10,
    holidays_in_sprint=0,
    contributors=[
        Contributor("Austin Sonderman", 1.0, 0),
        Contributor("Ben Digmann", 1.0, 0),
        Contributor("David Day", 1.0, 0),
        Contributor("George Mathews", 1.0, 0),
        Contributor("Nate Sepich", 1.0, 0),
        Contributor("Michael Moser", 1.0, 0),
        Contributor("Molly McCain", 1.0, 0),
        Contributor("Scott Lepech", 1.0, 0),
        Contributor("Tim Loungeway", 1.0, 0),
        Contributor("Senthil Vadivalagan Panchapakesan", 1.0, 0),
        Contributor("Vlad Orzhekhovskiy", 1.0, 0),
    ],
)

# --- HISTORICAL SPRINT DATA ---
# DEPRECATED: Sprint data now fetched from Jira via JiraClient
# This fallback data is kept for backward compatibility and testing
# See src/jira_client.py and get_historical_data_from_jira() in sprint_utils.py
HISTORICAL_SPRINTS = [
    # {
    #     "sprint": "Sprint 17",
    #     "committed": 89,
    #     "completed": 81,
    #     "carryover": 15,
    #     "sprint_length_days": 10,
    #     "holidays": 1,
    #     "contributors": [
    #         {"name": "Austin Sonderman", "commitment": 1.0, "pto_days": 0},
    #         {"name": "Ben Digmann", "commitment": 1.0, "pto_days": 0},
    #         {"name": "David Day", "commitment": 0.5, "pto_days": 0},
    #         {"name": "George Mathews", "commitment": 1.0, "pto_days": 0},
    #         {"name": "Nate Sepich", "commitment": 1.0, "pto_days": 0},
    #         {"name": "Michael Moser", "commitment": 1.0, "pto_days": 0},
    #         {"name": "Molly McCain", "commitment": 1.0, "pto_days": 0},
    #         {"name": "Scott Lepech", "commitment": 1.0, "pto_days": 0},
    #         {"name": "Tim Loungeway", "commitment": 1.0, "pto_days": 0},
    #         {"name": "Vlad Orzhekhovskiy", "commitment": 1.0, "pto_days": 0},
    #         {"name": "Keith Hopkins", "commitment": 0.1, "pto_days": 0},
    #     ],
    # },
    {
        "sprint": "Sprint 10",
        "committed": 57,
        "completed": 48,
        "carryover": 10,
        "sprint_length_days": 10,
        "holidays": 1,  # Memorial Day
        # Actual team capacity that sprint (effective person-days)
        "actual_capacity_days": 60.0,
        # Alternative: you can specify team composition if you have it
        "contributors": [
            {"name": "Austin Sonderman", "commitment": 1.0, "pto_days": 0},
            {"name": "Michael Moser", "commitment": 1.0, "pto_days": 0},
            {"name": "Rajesh Thakur", "commitment": 1.0, "pto_days": 0},
            {"name": "Rahul Verma", "commitment": 1.0, "pto_days": 0},
            {"name": "Scott Lepech", "commitment": 1.0, "pto_days": 0},
            {"name": "Vlad Orzhekhovskiy", "commitment": 1.0, "pto_days": 0},
        ],
    },
    {
        "sprint": "Sprint 11",
        "committed": 109,
        "completed": 78,
        "carryover": 26,
        "sprint_length_days": 10,
        "holidays": 0,
        "actual_capacity_days": 120.0,
        "contributors": [
            {"name": "Austin Sonderman", "commitment": 1.0, "pto_days": 0},
            {"name": "Ben Digmann", "commitment": 1.0, "pto_days": 0},
            {"name": "Brandon Campbell", "commitment": 1.0, "pto_days": 0},
            {"name": "David Day", "commitment": 0.5, "pto_days": 0},
            {"name": "Delbert Murphy", "commitment": 1.0, "pto_days": 0},
            {"name": "George Mathews", "commitment": 1.0, "pto_days": 0},
            {"name": "Michael Moser", "commitment": 1.0, "pto_days": 0},
            {"name": "Molly McCain", "commitment": 1.0, "pto_days": 0},
            {"name": "Rajesh Thakur", "commitment": 1.0, "pto_days": 0},
            {"name": "Rahul Verma", "commitment": 1.0, "pto_days": 0},
            {"name": "Scott Lepech", "commitment": 1.0, "pto_days": 0},
            {"name": "Tim Loungeway", "commitment": 1.0, "pto_days": 0},
            {
                "name": "Senthil Vadivalagan Panchapakesan",
                "commitment": 1.0,
                "pto_days": 0,
            },
            {"name": "Vlad Orzhekhovskiy", "commitment": 1.0, "pto_days": 0},
        ],
    },
    {
        "sprint": "Sprint 12",
        "committed": 136,
        "completed": 70,
        "carryover": 43,
        "sprint_length_days": 10,
        "holidays": 0,
        "actual_capacity_days": 125.0,
        "contributors": [
            {"name": "Austin Sonderman", "commitment": 1.0, "pto_days": 0},
            {"name": "Ben Digmann", "commitment": 1.0, "pto_days": 0},
            {"name": "Brandon Campbell", "commitment": 1.0, "pto_days": 0},
            {"name": "David Day", "commitment": 0.5, "pto_days": 0},
            {"name": "Delbert Murphy", "commitment": 1.0, "pto_days": 0},
            {"name": "George Mathews", "commitment": 1.0, "pto_days": 0},
            {"name": "Michael Moser", "commitment": 1.0, "pto_days": 0},
            {"name": "Molly McCain", "commitment": 1.0, "pto_days": 0},
            {"name": "Rajesh Thakur", "commitment": 1.0, "pto_days": 0},
            {"name": "Rahul Verma", "commitment": 1.0, "pto_days": 0},
            {"name": "Scott Lepech", "commitment": 1.0, "pto_days": 0},
            {"name": "Tim Loungeway", "commitment": 1.0, "pto_days": 0},
            {
                "name": "Senthil Vadivalagan Panchapakesan",
                "commitment": 1.0,
                "pto_days": 0,
            },
            {"name": "Vlad Orzhekhovskiy", "commitment": 1.0, "pto_days": 0},
        ],
    },
    {
        "sprint": "Sprint 13",
        "committed": 82,
        "completed": 71,
        "carryover": 11,
        "sprint_length_days": 10,
        "holidays": 1,  # July 4th
        "actual_capacity_days": 76.0,
        "contributors": [
            {"name": "Austin Sonderman", "commitment": 1.0, "pto_days": 0},
            {"name": "Ben Digmann", "commitment": 1.0, "pto_days": 0},
            {"name": "David Day", "commitment": 0.5, "pto_days": 0},
            {"name": "George Mathews", "commitment": 1.0, "pto_days": 0},
            {"name": "Michael Moser", "commitment": 1.0, "pto_days": 0},
            {"name": "Molly McCain", "commitment": 1.0, "pto_days": 0},
            {"name": "Rajesh Thakur", "commitment": 1.0, "pto_days": 0},
            {"name": "Scott Lepech", "commitment": 1.0, "pto_days": 0},
            {"name": "Tim Loungeway", "commitment": 1.0, "pto_days": 0},
            {
                "name": "Senthil Vadivalagan Panchapakesan",
                "commitment": 1.0,
                "pto_days": 0,
            },
            {"name": "Vlad Orzhekhovskiy", "commitment": 1.0, "pto_days": 0},
        ],
    },
    {
        "sprint": "Sprint 14",
        "committed": 79,
        "completed": 73,
        "carryover": 24,
        "sprint_length_days": 10,
        "holidays": 1,
        "contributors": [
            {"name": "Austin Sonderman", "commitment": 1.0, "pto_days": 0},
            {"name": "Ben Digmann", "commitment": 1.0, "pto_days": 0},
            {"name": "David Day", "commitment": 0.5, "pto_days": 0},
            {"name": "George Mathews", "commitment": 1.0, "pto_days": 0},
            {"name": "Nate Sepich", "commitment": 1.0, "pto_days": 0},
            {"name": "Michael Moser", "commitment": 1.0, "pto_days": 0},
            {"name": "Molly McCain", "commitment": 1.0, "pto_days": 0},
            {"name": "Scott Lepech", "commitment": 1.0, "pto_days": 0},
            {"name": "Tim Loungeway", "commitment": 1.0, "pto_days": 0},
            {
                "name": "Senthil Vadivalagan Panchapakesan",
                "commitment": 1.0,
                "pto_days": 0,
            },
            {"name": "Vlad Orzhekhovskiy", "commitment": 1.0, "pto_days": 0},
        ],
    },
    {
        "sprint": "Sprint 15",
        "committed": 74,
        "completed": 79,
        "carryover": 9,
        "sprint_length_days": 10,
        "holidays": 0,
        "contributors": [
            {"name": "Austin Sonderman", "commitment": 1.0, "pto_days": 0},
            {"name": "Ben Digmann", "commitment": 1.0, "pto_days": 0},
            {"name": "David Day", "commitment": 0.5, "pto_days": 0},
            {"name": "George Mathews", "commitment": 1.0, "pto_days": 4},
            {"name": "Nate Sepich", "commitment": 1.0, "pto_days": 0},
            {"name": "Michael Moser", "commitment": 1.0, "pto_days": 0},
            {"name": "Molly McCain", "commitment": 1.0, "pto_days": 1},
            {"name": "Scott Lepech", "commitment": 1.0, "pto_days": 0},
            {"name": "Tim Loungeway", "commitment": 1.0, "pto_days": 0},
            {
                "name": "Senthil Vadivalagan Panchapakesan",
                "commitment": 1.0,
                "pto_days": 0,
            },
            {"name": "Vlad Orzhekhovskiy", "commitment": 1.0, "pto_days": 0},
        ],
    },
    {
        "sprint": "Sprint 16",
        "committed": 83,
        "completed": 72,
        "carryover": 15,
        "sprint_length_days": 10,
        "holidays": 0,
        "contributors": [
            {"name": "Austin Sonderman", "commitment": 1.0, "pto_days": 0},
            {"name": "Ben Digmann", "commitment": 1.0, "pto_days": 0},
            {"name": "David Day", "commitment": 0.5, "pto_days": 0},
            {"name": "George Mathews", "commitment": 1.0, "pto_days": 2},
            {"name": "Nate Sepich", "commitment": 1.0, "pto_days": 2},
            {"name": "Michael Moser", "commitment": 1.0, "pto_days": 0},
            {"name": "Molly McCain", "commitment": 1.0, "pto_days": 1},
            {"name": "Scott Lepech", "commitment": 1.0, "pto_days": 0},
            {"name": "Tim Loungeway", "commitment": 1.0, "pto_days": 0},
            {"name": "Vlad Orzhekhovskiy", "commitment": 1.0, "pto_days": 0},
            {"name": "Brandon Wepking", "commitment": 1.0, "pto_days": 1},
        ],
    },
    {
        "sprint": "Sprint 17",
        "committed": 89,
        "completed": 87,
        "carryover": 9,
        "sprint_length_days": 10,
        "holidays": 1,
        "contributors": [
            {"name": "Austin Sonderman", "commitment": 1.0, "pto_days": 0},
            {"name": "Ben Digmann", "commitment": 1.0, "pto_days": 0},
            {"name": "David Day", "commitment": 0.5, "pto_days": 0},
            {"name": "George Mathews", "commitment": 1.0, "pto_days": 0},
            {"name": "Nate Sepich", "commitment": 1.0, "pto_days": 0},
            {"name": "Michael Moser", "commitment": 1.0, "pto_days": 0},
            {"name": "Molly McCain", "commitment": 1.0, "pto_days": 0},
            {"name": "Scott Lepech", "commitment": 1.0, "pto_days": 0},
            {"name": "Tim Loungeway", "commitment": 1.0, "pto_days": 0},
            {"name": "Vlad Orzhekhovskiy", "commitment": 1.0, "pto_days": 0},
            {"name": "Keith Hopkins", "commitment": 0.1, "pto_days": 2},
        ],
    },
]
