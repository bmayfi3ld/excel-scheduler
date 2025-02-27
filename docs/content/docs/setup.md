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
way needed, but here is an example of what the final result might look like:

| Search | PKA            |                |                |                |                |
| ------ | -------------- | -------------- | -------------- | -------------- | -------------- |
|        |                |                |                |                |                |
|        | Monday         | Tuesday        | Wednesday      | Thursday       | Friday         |
| 8am    | Gym            | PK3 Room       | -              | -              | -              |
| 9am    | PK3 Room       | Music          | -              | -              | -              |
| 10am   | -              | -              | -              | PE             | -              |
| 11am   | -              | -              | -              | -              | -              |
| 12pm   | Outside Lunch  | -              | -              | -              | -              |
| 1pm    | -              | -              | Art            | -              | -              |
| 2pm    | -              | -              | -              | -              | -              |
| 3pm    | -              | -              | -              | -              | -              |

The cells that show "-" indicate no class is scheduled for that cohort at that time.

To create this view, you would use the FINDCOHORTCLASS function in a formula like this in each cell of the schedule grid:

| Search | PKA            |                |                |                |                |
| ------ | -------------- | -------------- | -------------- | -------------- | -------------- |
|        |                |                |                |                |                |
|        | Monday         | Tuesday        | Wednesday      | Thursday       | Friday         |
| 8am    | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,B$3,$A4,Schedule!$A$1:$Z$50) | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,C$3,$A4,Schedule!$A$1:$Z$50) | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,D$3,$A4,Schedule!$A$1:$Z$50) | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,E$3,$A4,Schedule!$A$1:$Z$50) | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,F$3,$A4,Schedule!$A$1:$Z$50) |
| 9am    | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,B$3,$A5,Schedule!$A$1:$Z$50) | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,C$3,$A5,Schedule!$A$1:$Z$50) | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,D$3,$A5,Schedule!$A$1:$Z$50) | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,E$3,$A5,Schedule!$A$1:$Z$50) | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,F$3,$A5,Schedule!$A$1:$Z$50) |
| 10am   | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,B$3,$A6,Schedule!$A$1:$Z$50) | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,C$3,$A6,Schedule!$A$1:$Z$50) | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,D$3,$A6,Schedule!$A$1:$Z$50) | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,E$3,$A6,Schedule!$A$1:$Z$50) | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,F$3,$A6,Schedule!$A$1:$Z$50) |
| 11am   | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,B$3,$A7,Schedule!$A$1:$Z$50) | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,C$3,$A7,Schedule!$A$1:$Z$50) | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,D$3,$A7,Schedule!$A$1:$Z$50) | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,E$3,$A7,Schedule!$A$1:$Z$50) | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,F$3,$A7,Schedule!$A$1:$Z$50) |
| 12pm   | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,B$3,$A8,Schedule!$A$1:$Z$50) | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,C$3,$A8,Schedule!$A$1:$Z$50) | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,D$3,$A8,Schedule!$A$1:$Z$50) | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,E$3,$A8,Schedule!$A$1:$Z$50) | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,F$3,$A8,Schedule!$A$1:$Z$50) |
| 1pm    | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,B$3,$A9,Schedule!$A$1:$Z$50) | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,C$3,$A9,Schedule!$A$1:$Z$50) | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,D$3,$A9,Schedule!$A$1:$Z$50) | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,E$3,$A9,Schedule!$A$1:$Z$50) | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,F$3,$A9,Schedule!$A$1:$Z$50) |
| 2pm    | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,B$3,$A10,Schedule!$A$1:$Z$50) | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,C$3,$A10,Schedule!$A$1:$Z$50) | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,D$3,$A10,Schedule!$A$1:$Z$50) | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,E$3,$A10,Schedule!$A$1:$Z$50) | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,F$3,$A10,Schedule!$A$1:$Z$50) |
| 3pm    | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,B$3,$A11,Schedule!$A$1:$Z$50) | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,C$3,$A11,Schedule!$A$1:$Z$50) | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,D$3,$A11,Schedule!$A$1:$Z$50) | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,E$3,$A11,Schedule!$A$1:$Z$50) | =EXCELSCHEDULER.FINDCOHORTCLASS($B$1,F$3,$A11,Schedule!$A$1:$Z$50) |

With this setup, if you change the value in the cell next to "Search," it will
update the page with the latest schedule for the selected cohort.

To make selecting cohorts easier, you can set up Excel's data validation on the cell next to "Search" (cell B1 in this example):

1. Select cell B1
2. Go to Data → Data Validation
3. Choose "List" as the validation criteria
4. For the source, reference the range containing your cohort names from the "AllCohorts" rule
   (for example, =Rules!A2:A20)
5. Click OK

This will create a convenient dropdown menu in cell B1, allowing you to select from any of your defined cohorts. When you select a different cohort from this dropdown, all the formulas on the sheet will automatically update to show that cohort's schedule.

The FINDCOHORTCLASS function used in each cell references four pieces of information:

- The cohort name from cell B1 (which stays constant for all lookups)
- The day of the week from row 3 (which remains the same for each column)
- The time from column A (which stays the same for each row)
- The schedule range that includes all necessary data (the entire Schedule sheet in this example)

The dollar signs ($) in the formula lock specific cell references so they don't change when you copy the formula across the schedule grid. This single formula can be copied to all cells in the schedule, and it will automatically adjust to show the correct class for each time slot and day.

**Important Note:** Make sure to include the full Schedule range as the fourth parameter. The range should include the header row with day/time labels and all rows with class information.

For more details on how to use the provided functions, see the [Functions]({{< ref "functions" >}}) page.