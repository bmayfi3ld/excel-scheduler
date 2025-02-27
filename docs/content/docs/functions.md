+++
title = 'Functions'
weight = 40
+++
# Functions

Functions provide custom ways to retrieve data from the master schedule, which
can become large and not user-friendly to navigate directly.

## FINDCOHORTCLASS

The `FINDCOHORTCLASS` function searches for a class that a cohort is in given a day and time.

**Parameters**
- cohort: The name of the cohort you want to look up (e.g., "PKA", "K1B")
- day: The day of the week (Monday, Tuesday, Wednesday, Thursday, or Friday)
- timeslot: The time of day (e.g., "8am", "9am", "2pm")
- schedule: The range containing the schedule data
  - Must include the header row with day/time labels and the first column with the class names

**Return Value**
Returns the name of the class the cohort is scheduled for at the specified day and time. If no class is scheduled, returns "-".

**Example**
```
=CLASSSCHEDULER.FINDCOHORTCLASS("PKA", "Monday", "8am", Schedule!$A$1:$Z$50)
```
or when using cells as reference (see [Setup]({{< ref "setup" >}})) for an example
```
=CLASSSCHEDULER.FINDCOHORTCLASS($B$1,B$3,$A4,Schedule!A1:E5)
```

This will return "Gym" if PKA has gym class scheduled on Monday at 8am.

## Tips

- Make sure the cohort names and time periods match exactly what's in the master schedule
- If you reference cells for the parameters, make sure they contain values in the correct format
- The schedule parameter should include the entire schedule table, including headers
- Use absolute references (with $ signs) for the schedule range to avoid issues when copying formulas

## Common Issues

If you get an error, check that:

- The cohort name exists in the master schedule
- The day is spelled correctly
- The time format matches the expected format
- All four parameters are provided
- The schedule range includes both the header row with timeslot labels and all relevant rows with class information
- If you see "Array cannot be lifted over to call a function on individual array members" error, make sure you're passing the schedule as a range parameter, not as a table object