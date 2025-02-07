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

| AllCohorts | ClassRequiresTravel | *More Rules...* |
| ---------- | ------------------- | --------------- |
| 1st        | Latin               |                 |
| 2nd        |                     |                 |
| 3rd        | 1st                 |                 |
|            | 2nd                 |                 |
|            |                     |                 |
|            | 3rd                 |                 |


Often when you are configuring a rules in a single column, the rule with use `breaks`
to space out the different options. A break is just an empty cell. Notice above
ther is a `break` between `Latin` and `1st`.

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


## ClassRequiresTravel

Certain cohorts can not have certain classes sequentially. This could be due to setup
and teardown requirements or if a teacher has to physically travel to the cohort
and they are far apart.

eg: If there is a class called `Lunch Cart` or maybe `Latin Cart` or `Homeroom Art`
where the class physically travels to the homeroom of the cohort to teach the
class. The class cannot be taught to certain cohorts sequentially if they are in different
buildings.

**Configuration**:

A repeating pattern of `Class Name`, `break`, `all cohorts that located near
each other`, `break`, `another group of cohorts located near each other` and then
the pattern of `group of cohorts` `break` can continue for as many groups as you
want to setup for this class (perhaps you have 4 buildings, then you could have 4 groups).

Then add a `double break` and you can add another class.

eg:

Here we have two classes that need travel spacing `Latin Cart` which will be a
cart that has to physcially travel to the homeroom of the classes and `Lunch Cart`
which also has to travel. You will notice the double break between the two classes.

We can notice that it looks like 1st, 2nd, 3rd are in the same building as well as
4th, 5th, 6th in one and 7th, 8th, 9th in  another.

7th, 8th, 9th do not take the "Latin Cart" class in this example, so they are
not included in the latin cart options.

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


So here we would get one error if the schedule looked like this.


|            | Monday, 9am | Monday, 10am | Monday, 11am | Monday, 12pm | ... |
| ---------- | ----------- | ------------ | ------------ | ------------ | --- |
| Latin Cart | 1st         | 4th          |              |              |     |
| Lunch Cart |             |              | 1st          | 2nd          |     |

We get an error for `Latin Cart` at `Monday, 10am` because the cart cannot travel
from wherever 1st is located to wherever 4th is located in the transition time.

We do not get an error for the lunch cart because 1st and 2nd homeroom are close
enough to each other.