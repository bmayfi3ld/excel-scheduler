/* global Excel */

/**
 * Find the class that a cohort is in given a day and time.
 * @customfunction
 * @param {string} cohort The cohort to search for
 * @param {string} day The day of the week
 * @param {string} timeslot The timeslot
 * @returns {string} The class the cohort is in.
 */
async function FindCohortClass(cohort, day, timeslot) {
  // console.log("find cohort")
  try {
    return await Excel.run(async (context) => {
      const sheet = context.workbook.worksheets.getItem("Schedule");
      const range = sheet.getUsedRange();
      range.load("values");
      await context.sync();

      const values = range.values;
      const headerRow = values[0];

      // Find column with matching day,timeslot
      const columnIndex = headerRow.findIndex(
        header => header === `${day}, ${timeslot}`
      );

      if (columnIndex === -1) return "Time not found";

      // Find row with matching cohort name
      for (let i = 1; i < values.length; i++) {
        // console.log(`checking ${values[i][0]} for ${cohort}, from column ${columnIndex}`)
        if (values[i][columnIndex] === cohort) {
          return values[i][0];
        }
      }

      return "No Class";
    });
  } catch (error) {
    return "Error: " + error.message;
  }
}