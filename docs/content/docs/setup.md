+++
title = 'Setup A New Workbook'
weight = 20
+++
# Setup A New Workbook
Just need two sheets, `Rules` and `Schedule`.


## Schedule Sheet

The schedule sheet will be a single table with timeslots across the first row
and classes as the first column.

The timeslots must be in the format `Day, Hour` for the function `FindCohortClass` to
work.

The sheet must be named `Schedule`

eg:

|                   | Monday, 9am | Monday, 10am | Monday, 11am | Monday, 12pm | *More Times..* |
| ----------------- | ----------- | ------------ | ------------ | ------------ | -------------- |
| Latin Cart        | 1st         | 4th          |              |              |                |
| Lunch Cart        |             |              | 1st          | 2nd          |                |
| Preschool Art     |             |              | 3rd          | 4th          |                |
| PE Gym Section A  |             |              | 1st          | 2nd          |                |
| *More Classes...* |             |              | 1st          | 2nd          |                |

A class is anywhere (or anything) that a single cohort would want to do/go
to/participate in. This is not necesarily just a list of "classes" or "specials"

Here are some examples:

- Pre-school Art
- PE in Big Gym A
- PE in Big Gym B (if two cohorts take gym at the same time, just make two "sections" for that class)
- Latin Cart (maybe not a physical location, but the cart travels to the homeroom of the cohort)

## Rules Sheet

The rules sheet will have the list of rules to apply to the schedule. Each rule
will have its own column and will be configured with the rows below it.

This sheet must be named `Rules`

eg:

| AllCohorts | ClassRequiresTravel | *More Rules...* |
| ---------- | ------------------- | --------------- |
| 1st        | Latin               |                 |
| 2nd        |                     |                 |
| 3rd        | 1st                 |                 |
|            | 2nd                 |                 |
|            |                     |                 |
|            | 3rd                 |                 |

Check the [Scheduler Rules]({{< ref "rules" >}}) page for all available rules.

## Optional Classroom Sheet

This is a flexible sheet that might take advantage of the ....


TODO make function and docs