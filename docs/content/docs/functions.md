+++
title = 'Functions'
weight = 40
+++
# Functions

Functions provide custom ways to retrieve data from the master schedule, which
can become large and not user-friendly to navigate directly.

## FINDCOHORTCLASS

The `FINDCOHORTCLASS` function searches for a class that a cohort is in given a day and time

**Parameters**
- cohort: The name of the cohort you want to look up (e.g., "PKA", "K1B")
- day: The day of the week (Monday, Tuesday, Wednesday, Thursday, or Friday)
- timeslot: The time of day (e.g., "8am", "9am", "2pm")

**Return Value**
Returns the name of the class the cohort is scheduled for at the specified day and time. If no class is scheduled, returns "No Class".

**Example**
```
=FINDCOHORTCLASS("PKA", "Monday", "8am")
```

This will return "Gym" if PKA has gym class scheduled on Monday at 8am.

## Tips

- Make sure the cohort names and time periods match exactly what's in the master schedule
- If you reference cells for the parameters, make sure they contain values in the correct format

## Common Issues

If you get an error, check that:

- The cohort name exists in the master schedule
- The day is spelled correctly
- The time format matches the expected format
- All three parameters are provided