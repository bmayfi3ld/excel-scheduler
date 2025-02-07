+++
title = 'Scheduler Rules'
weight = 30
+++
# Scheduler Rules

There are a number of rules you can use, they are listed here along with how
to configure them on the `Rules` sheet. When a rule has been broken the cell
will be highlighted `Red` and a comment will note the infraction reasons.

Rules are added by adding the rule name to the first row of the `Rules` sheet.
Each column will specifcy a new rule along with its options in the cell under
the name.

eg:

| AllCohorts | DestinationRequiresTravel | *More Rules...* |
| ---------- | ------------------------- | ----------- |
| 1st        | Latin                     |             |
| 2nd        |                           |             |
| 3rd        | 1st                       |             |
|            | 2nd                       |             |
|            |                           |             |
|            | 3rd                       |             |

## AllCohorts
Require each cohort on the schedule to be a valid cohort from this AllCohorts list.

eg: If you put 53rd into the schedule and that isn't in the list, it will be
marked.

**Configuration**:

Just a list of all possible cohorts.

eg:


| AllClasses |
| ---------- |
| 1st        |
| 2nd        |
| 3rd        |


## DestinationRequiresSetup

These destinations are in separate buildings