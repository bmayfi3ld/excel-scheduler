/* global console, document, Excel, Office */

Office.onReady((info) => {
  if (info.host === Office.HostType.Excel) {
    document.getElementById("sideload-msg").style.display = "none";
    document.getElementById("app-body").style.display = "flex";
    document.getElementById("run").onclick = run;
    document.getElementById("clear").onclick = clear;
  }
});

export async function run() {
  // Get the icon element
  const icon = document.querySelector(".ms-Icon");
  console.log("Starting validation script");

  await clear();

  try {
    // Log icon state
    console.log("Current icon classes:", icon ? icon.className : "Icon element not found");

    // Change to loading icon if element exists
    if (icon) {
      icon.classList.remove("ms-Icon--Ribbon");
      icon.classList.add("ms-Icon--Sync", "loading");
      console.log("Updated icon classes:", icon.className);
    } else {
      console.warn("Icon element not found");
    }

    await Excel.run(async (context) => {
      console.log("Starting Excel.run");

      // Get the Schedule sheet
      const scheduleSheet = context.workbook.worksheets.getItem("Schedule");
      console.log("Got Schedule sheet");

      const scheduleRange = scheduleSheet.getUsedRange();
      scheduleRange.load(["values", "rowCount", "columnCount"]);
      await context.sync();

      console.log(`Sheet dimensions: ${scheduleRange.rowCount} rows, ${scheduleRange.columnCount} columns`);

      // get the rules sheet
      const rulesSheet = context.workbook.worksheets.getItem("Rules");

      const rulesRange = rulesSheet.getUsedRange();
      rulesRange.load("values");
      await context.sync();

      // Get Values and Set Note for Rules
      const allCohortsConfig = await getValuesFromSheet(context, "AllCohorts", rulesRange);
      console.log("allCohortsConfig:", allCohortsConfig);
      addValidationToColumn(
        context,
        "AllCohorts",
        "AllCohorts",
        "A list of all the cohorts. eg: 1st, 2nd, 3rd",
        rulesRange
      );

      const classRequiresTravelConfig = await getValuesFromSheet(context, "ClassRequiresTravel", rulesRange);
      const classRequiresTravelParsed = splitArrayByEmptyStrings(classRequiresTravelConfig.values);
      console.log("classRequiresTravelConfig:", classRequiresTravelParsed);
      addValidationToColumn(
        context,
        "ClassRequiresTravel",
        "ClassRequiresTravel",
        "For a given class the following groups of cohorts cannot take the class sequentially, due to travel or other time restrictions.",
        rulesRange
      );

      // Iterate through each cell, starting from row 2 (skip header) and column 2 (skip first column)
      for (let row = 1; row < scheduleRange.rowCount; row++) {
        const className = scheduleRange.values[row][0];
        console.log(`checking timeslots for class ${className}`);

        for (let col = 1; col < scheduleRange.columnCount; col++) {
          const cellValue = scheduleRange.values[row][col];
          console.log(`Checking cell at ${getColumnLetter(col)}${row + 1} :`, {
            value: cellValue,
            isEmpty: !cellValue,
          });

          const priorCellValue = scheduleRange.values[row][col - 1];
          console.log(`prior class ${priorCellValue}`);

          // Skip empty cells
          if (!cellValue) {
            console.log("Skipping empty cell");
            continue;
          }

          let brokenRules = [];

          // Check AllCohortsRule
          if (!allCohortsConfig.values.includes(cellValue)) {
            console.log(`Invalid cohort found: "${cellValue}"`);

            brokenRules.push(
              "This class isn't in the total list of classes, check column " +
                allCohortsConfig.column +
                " on the Rules sheet."
            );
          }

          // Check ClassRequiresTravel
          classRequiresTravelParsed.forEach((classTravelConfig) => {
            // individual classtravelconfig pattern
            // 0: class name
            // 1: list of classes in building 1
            // 2: list of classes in next building
            // .... repeat

            if (classTravelConfig.length < 3) {
              console.log("Skipping class config, not enough parameters");
              // need at least the class name and 2 buildings
              return;
            }

            // if it isn't our class skip
            if (classTravelConfig[0] != className) {
              return;
            }

            let foundClassBuilding = -1;

            // Find which building contains the current class
            for (let i = 1; i < classTravelConfig.length; i++) {
              if (classTravelConfig[i].includes(cellValue)) {
                foundClassBuilding = i;
                break;
              }
            }

            // If class was found, check if prior class exists in any other building
            if (foundClassBuilding !== -1) {
              for (let i = 1; i < classTravelConfig.length; i++) {
                if (i !== foundClassBuilding && classTravelConfig[i].includes(priorCellValue)) {
                  brokenRules.push(
                    `The class ${className} can't go to one cohort ${cellValue} if the previous one was ${priorCellValue}, it is too far away (or requires setup) see column ${classRequiresTravelConfig.column} on the Rules sheet`
                  );
                  break;
                }
              }
            }
          });

          // add errors
          if (brokenRules.length != 0) {
            console.log(`adding broken rules ${brokenRules}`);

            // Get the specific cell
            const cell = scheduleRange.getCell(row, col);

            // Set fill color to red
            cell.format.fill.color = "red";

            brokenRules.forEach((comment) => {
              try {
                // Add comment using Excel.Comment
                scheduleSheet.comments.add(cell, comment);
                console.log(`Added comment ${comment} successfully`);
              } catch (commentError) {
                console.error("Error adding comment:", commentError);
              }
            });
          }
        }
      }

      console.log("Completing final context.sync()");
      await context.sync();
      console.log("Excel operations completed successfully");
    });
  } catch (error) {
    console.error("Error in validation script:", error);
  } finally {
    // Restore original icon
    if (icon) {
      icon.classList.remove("ms-Icon--Sync", "loading");
      icon.classList.add("ms-Icon--Ribbon");
      console.log("Restored icon to original state:", icon.className);
    }
    console.log("Script execution completed");
  }
}

// Helper function to convert column index to letter
function getColumnLetter(columnIndex) {
  let temp = columnIndex;
  let letter = "";

  while (temp >= 0) {
    letter = String.fromCharCode((temp % 26) + 65) + letter;
    temp = Math.floor(temp / 26) - 1;
  }

  return letter;
}

/**
 * add a validation tooltip to a the found cell
 * @param {Excel.RequestContext} context
 * @param {string} headerValue
 * @param {Excel.Range} rulesRange
 */
async function getValuesFromSheet(context, headerValue, rulesRange) {
  try {
    let startRow = -1;
    const values = rulesRange.values;

    // Find the row with the header value
    for (let i = 0; i < values.length; i++) {
      if (values[i].includes(headerValue)) {
        startRow = i;
        break;
      }
    }

    if (startRow === -1) {
      throw new Error(`Header '${headerValue}' not found`);
    }

    // Get the column index where header was found
    const columnIndex = values[startRow].indexOf(headerValue);

    // Convert column index to letter
    const columnLetter = getColumnLetter(columnIndex);

    // Create array of values (excluding the header)
    const resultArray = values
      .slice(startRow + 1) // Start from next row after header
      .map((row) => row[columnIndex]); // Get value from the same column

    return {
      values: resultArray,
      column: columnLetter,
    };
  } catch (error) {
    console.error("Error: ", error);
    throw error;
  }
}

/**
 * add a validation tooltip to a the found cell
 * @param {Excel.RequestContext} context
 * @param {string} headerValue
 * @param {string} validationTitle
 * @param {string} validationMessage
 * @param {Excel.Range} rulesRange
 */
async function addValidationToColumn(context, headerValue, validationTitle, validationMessage, rulesRange) {
  try {
    const values = rulesRange.values;

    // Get the column index where header was found
    const columnIndex = values[0].indexOf(headerValue);

    console.log("adding validation to row 0 and column " + columnIndex);
    const rulesSheet = context.workbook.worksheets.getItem("Rules");
    const validationRange = rulesSheet.getCell(0, columnIndex);

    // Add data validation
    validationRange.dataValidation.clear();

    validationRange.dataValidation.prompt = {
      message: validationMessage,
      showPrompt: true,
      title: validationTitle,
    };

    await context.sync();
  } catch (error) {
    console.error("Error: ", error);
  }
}

export async function clear() {
  try {
    await Excel.run(async (context) => {
      console.log("Starting Excel.clear");

      // Get the Schedule sheet
      const sheet = context.workbook.worksheets.getItem("Schedule");
      console.log("Got Schedule sheet");

      // Get the used range of the sheet
      const usedRange = sheet.getUsedRange();
      usedRange.load(["values", "rowCount", "columnCount"]);
      await context.sync();

      console.log(`Sheet dimensions: ${usedRange.rowCount} rows, ${usedRange.columnCount} columns`);
      console.log("Full range values:", usedRange.values);

      // Iterate through each cell, starting from row 2 (skip header) and column 2 (skip first column)
      for (let row = 1; row < usedRange.rowCount; row++) {
        for (let col = 1; col < usedRange.columnCount; col++) {
          const cellValue = usedRange.values[row][col];

          // Skip empty cells
          if (!cellValue) {
            console.log("Skipping empty cell");
            continue;
          }

          // Get the specific cell
          const cell = usedRange.getCell(row, col);
          console.log("Got cell reference");

          // Set fill color to none, cleaning up
          cell.format.fill.clear();
        }
      }

      sheet.load(["comments"]);
      await context.sync();

      sheet.comments.items.forEach((comment) => {
        comment.delete();
      });

      console.log("Completing final context.sync()");
      await context.sync();
      console.log("Excel operations completed successfully");
    });
  } catch (error) {
    console.error("Error in clear script:", error);
  }
}

/**
 * Splits an array into nested subarrays when empty strings are encountered.
 *
 * @param {Array} arr - The input array to be split
 * @returns {Array} Nested array with groups separated by empty strings
 *
 * @example
 * // Returns [[['a', 'b'], ['c', 'd']], [['e', 'f']]]
 * splitArrayByEmptyStrings(['a', 'b', '', 'c', 'd', '', '', 'e', 'f'])
 */
function splitArrayByEmptyStrings(arr) {
  const result = [];
  let currentGroup = [];
  let currentSubarray = [];
  let emptyCount = 0;

  for (const item of arr) {
    if (item === "") {
      emptyCount++;

      if (currentSubarray.length > 0) {
        currentGroup.push(currentSubarray);
        currentSubarray = [];
      }

      if (emptyCount === 2) {
        result.push(currentGroup);
        currentGroup = [];
        emptyCount = 0;
      }
    } else {
      currentSubarray.push(item);
      emptyCount = 0;
    }
  }

  if (currentSubarray.length > 0) {
    currentGroup.push(currentSubarray);
  }

  if (currentGroup.length > 0) {
    result.push(currentGroup);
  }

  return result;
}
