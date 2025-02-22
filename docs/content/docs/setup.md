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
- Timeslots must be in the format "Day, Period" for the `FINDCOHORTCLASS` function to work
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

There are functions that will help set up additional sheets that can be used
as views, providing more user-friendly ways to read the schedule.

### Cohort Schedule

To view a schedule for a single cohort in a more traditional calendar format,
the FINDCOHORTCLASS function can be used. This sheet can be customized in any
way needed, but here is an example of how it might be set up.

| Search | PKA            |                |                |                |                |
| ------ | -------------- | -------------- | -------------- | -------------- | -------------- |
|        |                |                |                |                |                |
|        | Monday         | Tuesday        | Wednesday      | Thursday       | Friday         |
| 8am    | Gym            | PK3 Room       | No Class | No Class | No Class |
| 9am    | PK3 Room       | Music          | No Class | No Class | No Class |
| 10am   | No Class | No Class | No Class | PE             | No Class |
| 11am   | No Class | No Class | No Class | No Class | No Class |
| 12pm   | Outside Lunch  | No Class | No Class | No Class | No Class |
| 1pm    | No Class | No Class | Art            | No Class | No Class |
| 2pm    | No Class | No Class | No Class | No Class | No Class |
| 3pm    | No Class | No Class | No Class | No Class | No Class |

With this setup, if you change the value in the cell next to "Search," it will
update the page with the latest schedule for the selected cohort. If data
validation is set up for that cell using the values from the "AllCohorts" rule,
there will be a dropdown menu that lets you select the desired cohort.

The formula =CLASSSCHEDULER.FINDCOHORTCLASS($B$1,B$3,$A4) is used in each schedule cell to look up the class.

It references three pieces of information:

- The cohort name from cell B1 (which stays constant for all lookups)
- The day of the week from row 3 (which remains the same for each column)
- and the time from column A (which stays the same for each row).

The dollar signs ($) in the formula lock specific cell references so they don't change when you copy the formula across the schedule grid. This single formula can be copied to all cells in the schedule, and it will automatically adjust to show the correct class for each time slot and day.

For more details on how to use the provided functions, see the [Functions]({{< ref "functions" >}}) page.