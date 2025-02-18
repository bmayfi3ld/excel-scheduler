+++
title = 'Scheduler Rules'
weight = 30
+++
# Scheduler Rules

Rules help validate your schedule and are located on the `Rules` sheet. When you break a rule, the relevant cell
will be highlighted in `Red` with a comment explaining the issue.

## Adding Rules

Add rules by putting the rule name in the first row of the Rules sheet. Each
column represents a new rule with its options listed below the name.

Example:

| AllCohorts | ClassRequiresTravel | *More Rules...* |
| ---------- | ------------------- | --------------- |
| 1st        | Latin               |                 |
| 2nd        |                     |                 |
| 3rd        | 1st                 |                 |
|            | 2nd                 |                 |
|            |                     |                 |
|            | 3rd                 |                 |

**Two Important Notes:**
- Rules must be named exactly as they are in the documentation.
- Names of classes and cohorts must be consistent.

When configuring rules, empty cells will be used as `breaks` to separate different
options. Notice the `break` (empty cell) between Latin and 1st above.

## Available Rules Reference

### AllCohorts
This rule ensures that every cohort in your schedule matches an entry in your
valid cohort list. For example, if you enter "53rd" in the schedule but it's
not in your list, it will be marked as an error.

**Configuration**:

List all possible cohorts.

Example:


| AllClasses |
| ---------- |
| 1st        |
| 2nd        |
| 3rd        |


### ClassRequiresTravel

This rule prevents certain cohorts from having specific classes back-to-back.

Use it when:
- Classes need setup or teardown time
- Teachers must travel between cohorts
- Mobile classes (like "Lunch Cart" or "Latin Cart") move between different buildings

**Configuration**:

Follow this pattern:

1. Class Name
2. Break (empty cell)
3. Group of cohorts located near each other, each in its own cell
4. Break
5. Another group of colocated cohorts
6. Continue with breaks and cohort groups as needed
7. Double break before starting a new class

Example:

| AllClasses |
| ---------- |
| Latin Cart |
|            |
| 1st        |
| 2nd        |
| 3rd        |
|            |
| 4th        |
| 5th        |
| 6th        |
|            |
|            |
| Lunch Cart |
|            |
| 1st        |
| 2nd        |
| 3rd        |
|            |
| 4th        |
| 5th        |
| 6th        |
|            |
| 7th        |
| 8th        |
| 9th        |

Here we have two classes that need to travel to each cohorts homeroom `Latin Cart`
 and `Lunch Cart`.

This example shows:
- 1st, 2nd, 3rd are in the same building
- 4th, 5th, 6th are in another
- and 7th, 8th, 9th in yet another

7th, 8th, 9th do not take the `Latin Cart` class in this example, so they are
not included in the latin cart options.


Here is an example schedule that would give one error:


|            | Monday, 9am | Monday, 10am | Monday, 11am | Monday, 12pm | ... |
| ---------- | ----------- | ------------ | ------------ | ------------ | --- |
| Latin Cart | 1st         | 4th          |              |              |     |
| Lunch Cart |             |              | 1st          | 2nd          |     |

We get an error for `Latin Cart` at `Monday, 10am` because the cart cannot travel
from wherever 1st is located to wherever 4th is located in the transition time.

We do not get an error for the `Lunch Cart` because 1st and 2nd homeroom are
located close to each other.