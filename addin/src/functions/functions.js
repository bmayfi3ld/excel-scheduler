/**
 * Find the class that a cohort is in given a day and time.
 * @customfunction
 * @param {string} cohort The cohort to search for
 * @param {string} day The day of the week
 * @param {string} timeslot The timeslot
 * @param {string[][]} schedule The entire schedule table range
 * @returns {string} The class the cohort is in.
 */
function FindCohortClass(cohort, day, timeslot, schedule) {
  try {
    // Format the column header we're looking for
    const headerToFind = `${day}, ${timeslot}`;

    // First row contains headers
    const headerRow = schedule[0];

    // Find the column index with the matching day/timeslot header
    const columnIndex = headerRow.findIndex(header => header === headerToFind);
    if (columnIndex === -1) return "error: time or day not found";

    // Search each row for the cohort in the target column
    for (let i = 1; i < schedule.length; i++) {
      const rowData = schedule[i];

      // If this row has the cohort in the target column
      if (rowData[columnIndex] === cohort) {
        // Return the class name (from the first column)
        return rowData[0];
      }
    }

    // If cohort wasn't found in that timeslot
    return "-";
  } catch (error) {
    return "Error: " + error.message;
  }
}