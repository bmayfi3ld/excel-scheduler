+++
title = 'Setup A New Workbook'
weight = 20
+++
# Setup A New Workbook

For a workbook there are only two specific sheets required to use this
app: `Rules` and `Schedule`.

## Schedule Sheet

The `Schedule` sheet contains a single table with timeslots across the first row
and classes in the first column. The table is then filled with cohorts. A cohort is
any group of students or individuals that would want to participate in a class
at a time.

Important requirements:
- Timeslots must be in the format Day, Period for the `FindCohortClass` [In Development] function to work
- The sheet must be named `Schedule`

Example:

|                   | Monday, 9am | Monday, 10am | Monday, 11am-11:45am | Monday, 12pm | *More Times..* |
| ----------------- | ----------- | ------------ | ------------ | ------------ | -------------- |
| Latin Cart        | 1st         | 4th          |              |              |                |
| Lunch Cart        |             |              | 1st          | 2nd          |                |
| Preschool Art     |             |              | 3rd          | 4th          |                |
| PE Gym Section A  |             |              | 1st          | 2nd          |                |
| *More Classes...* |             |              | 1st          | 2nd          |                |

The time period can be in any format. Example: 9am or 9:45am-10:45am.

A class represents any activity or location where a single cohort would participate. This isn't limited to traditional classes or specials. Examples include:

- Pre-school Art
- PE in Big Gym A
- PE in Big Gym B (create separate sections for multiple cohorts in the same space)
- Latin Cart (mobile activities that travel to cohort homerooms)

## Rules Sheet

The `Rules` sheet contains your list of scheduling rules. Each rule has its own
column and is configured using the rows below it.

- This sheet must be named `Rules`

Example:

| AllCohorts | ClassRequiresTravel | *More Rules...* |
| ---------- | ------------------- | --------------- |
| 1st        | Latin               |                 |
| 2nd        |                     |                 |
| 3rd        | 1st                 |                 |
|            | 2nd                 |                 |
|            |                     |                 |
|            | 3rd                 |                 |

For a complete list of available rules, see the [Scheduler Rules]({{< ref "rules" >}}) page.

## Other Sheets

[In Development]
There are functions that will help setup additional sheets that can be used
as views, to provide more friendly ways to read the schedule.