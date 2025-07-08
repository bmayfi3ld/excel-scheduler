/* global console, document, Excel, Office */

Office.onReady((info) => {
  if (info.host === Office.HostType.Excel) {
    document.getElementById("sideload-msg").style.display = "none";
    document.getElementById("app-body").style.display = "flex";

    // Set up toggle event handler
    const toggleCheckbox = document.getElementById("auto-check-toggle");
    toggleCheckbox.addEventListener("change", handleToggleChange);

    // Initialize event handlers for worksheet changes
    setupWorksheetChangeHandlers();
  }
});

// Track if auto-check is enabled
let autoCheckEnabled = false;
let changeHandler = null;

// Handle toggle change
function handleToggleChange(event) {
  autoCheckEnabled = event.target.checked;

  if (autoCheckEnabled) {
    // If enabled, set up event handlers and run rules check
    setupWorksheetChangeHandlers();
    run();
  } else {
    // If disabled, remove event handlers and clear rules
    removeWorksheetChangeHandlers();
    clear();
  }
}

// Set up handlers to watch for changes in the Rules and Schedule sheets
async function setupWorksheetChangeHandlers() {
  try {
    await Excel.run(async (context) => {
      // Get the Rules and Schedule sheets
      const rulesSheet = context.workbook.worksheets.getItem("Rules");
      const scheduleSheet = context.workbook.worksheets.getItem("Schedule");

      // Register the onChange event handler for both sheets
      rulesSheet.onChanged.add(handleWorksheetChange);
      scheduleSheet.onChanged.add(handleWorksheetChange);

      await context.sync();
      console.log("excel-scheduler worksheet change handlers set up successfully");
    });
  } catch (error) {
    console.error("excel-scheduler error setting up worksheet change handlers:", error);
  }
}

// Remove the worksheet change handlers
async function removeWorksheetChangeHandlers() {
  try {
    await Excel.run(async (context) => {
      // Get the Rules and Schedule sheets
      const rulesSheet = context.workbook.worksheets.getItem("Rules");
      const scheduleSheet = context.workbook.worksheets.getItem("Schedule");

      // Remove the onChange event handlers
      rulesSheet.onChanged.remove();
      scheduleSheet.onChanged.remove();

      await context.sync();
      console.log("excel-scheduler worksheet change handlers removed successfully");
    });
  } catch (error) {
    console.error("excel-scheduler error removing worksheet change handlers:", error);
  }
}

// Handle worksheet changes
async function handleWorksheetChange(event) {
  // Only process if auto-check is enabled
  if (autoCheckEnabled) {
    console.log("excel-scheduler worksheet changed, running rules check");

    // Debounce the rule check to avoid running it too frequently
    if (changeHandler) {
      clearTimeout(changeHandler);
    }

    // Wait a short delay before running the check to batch multiple changes
    changeHandler = setTimeout(async () => {
      await run();
      changeHandler = null;
    }, 3000); // 1 second debounce
  }
}

export async function run() {
  // Start timer
  const startTime = performance.now();

  // Get the icon element
  const icon = document.querySelector(".ms-Icon");
  console.log("excel-scheduler starting validation script");

  await clear();

  try {
    await Excel.run(async (context) => {
      console.log("excel-scheduler starting rule check");

      // Get the Schedule sheet
      const scheduleSheet = context.workbook.worksheets.getItem("Schedule");

      const scheduleRange = scheduleSheet.getUsedRange();
      scheduleRange.load(["values", "rowCount", "columnCount"]);

      // get the rules sheet
      const rulesSheet = context.workbook.worksheets.getItem("Rules");

      const rulesRange = rulesSheet.getUsedRange();
      rulesRange.load("values");
      await context.sync();

      console.log(
        `excel-scheduler sheet dimensions: ${scheduleRange.rowCount} rows, ${scheduleRange.columnCount} columns`
      );

      // Get Values and Set Note for Rules
      const allCohortsConfig = await getValuesFromSheet(context, "AllCohorts", rulesRange);
      console.log("excel-scheduler allCohortsConfig:", allCohortsConfig);
      addValidationToColumn(
        context,
        "AllCohorts",
        "AllCohorts",
        "A list of all the cohorts. eg: 1st, 2nd, 3rd",
        rulesRange
      );

      const classRequiresTravelConfig = await getValuesFromSheet(context, "ClassRequiresTravel", rulesRange);
      const classRequiresTravelParsed = splitArrayByEmptyStrings(classRequiresTravelConfig.values);
      console.log("excel-scheduler classRequiresTravelConfig:", classRequiresTravelParsed);
      addValidationToColumn(
        context,
        "ClassRequiresTravel",
        "ClassRequiresTravel",
        "For a given class the following groups of cohorts cannot take the class sequentially, due to travel or other time restrictions.",
        rulesRange
      );

      const cohortBlacklistConfig = await getValuesFromSheet(context, "CohortBlacklist", rulesRange);
      const cohortBlacklistParsed = splitArrayByEmptyStrings(cohortBlacklistConfig.values);
      console.log("excel-scheduler cohortBlacklistConfig:", cohortBlacklistParsed);
      addValidationToColumn(
        context,
        "CohortBlacklist",
        "CohortBlacklist",
        "Defines time slots when specific cohorts are not allowed to have any classes (e.g., lunch periods, breaks).",
        rulesRange
      );

      // Check for OneClassAtATime rule
      let oneClassAtATimeEnabled = false;
      try {
        const oneClassAtATimeConfig = await getValuesFromSheet(context, "OneClassAtATime", rulesRange);
        if (oneClassAtATimeConfig && oneClassAtATimeConfig.values.length > 0) {
          // This rule is enabled if it exists in the Rules sheet
          oneClassAtATimeEnabled = true;
          console.log("excel-scheduler OneClassAtATime rule is enabled");

          addValidationToColumn(
            context,
            "OneClassAtATime",
            "OneClassAtATime",
            "When enabled, ensures cohorts are not scheduled for multiple classes during the same time slot.",
            rulesRange
          );
        }
      } catch (error) {
        console.log("excel-scheduler OneClassAtATime rule not found or invalid", error);
      }

      // Iterate through each cell, starting from row 2 (skip header) and column 2 (skip first column)
      for (let row = 1; row < scheduleRange.rowCount; row++) {
        const className = scheduleRange.values[row][0];
        console.log(`excel-scheduler checking timeslots for class ${className}`);

        for (let col = 1; col < scheduleRange.columnCount; col++) {
          const cellValue = scheduleRange.values[row][col];
          if (!cellValue) {
            continue;
          }

          // Skip rule checking if cell has 4 or more repeating characters in first 4 characters
          if (cellValue.length >= 4) {
            const first4Chars = cellValue.substring(0, 4);
            const firstChar = first4Chars.charAt(0);
            let repeatingCount = 0;

            for (let i = 0; i < first4Chars.length; i++) {
              if (first4Chars.charAt(i) === firstChar) {
                repeatingCount++;
              }
            }

            if (repeatingCount >= 4) {
              console.log(`excel-scheduler skipping rule check for cell with repeating characters: "${cellValue}"`);
              continue;
            }
          }

          console.log(`excel-scheduler checking cell at ${getColumnLetter(col)}${row + 1} :`, {
            value: cellValue,
            isEmpty: !cellValue,
          });

          const priorCellValue = scheduleRange.values[row][col - 1];
          console.log(`excel-scheduler prior class ${priorCellValue}`);

          let brokenRules = [];

          // Check AllCohortsRule
          if (!allCohortsConfig.values.includes(cellValue)) {
            console.log(`excel-scheduler invalid cohort found: "${cellValue}"`);

            brokenRules.push(
              `The cohort '${cellValue}' isn't in the total list of classes, check column '${allCohortsConfig.column}' on the Rules sheet.`
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
              console.log("excel-scheduler skipping class config, not enough parameters");
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
                    `The class '${className}' can't go to one cohort '${cellValue}' if the previous one was '${priorCellValue}', it is too far away (or requires setup) see column '${classRequiresTravelConfig.column}' on the Rules sheet`
                  );
                  break;
                }
              }
            }
          });

          // Check CohortBlacklist
          cohortBlacklistParsed.forEach((blacklistConfig) => {
            // individual blacklistConfig pattern:
            // 0: cohort name
            // 1: list of blacklisted timeslots

            if (blacklistConfig.length < 2) {
              console.log(`excel-scheduler skipping blacklist config, not enough parameters, got ${blacklistConfig}`);
              return;
            }

            const cohortName = blacklistConfig[0][0];

            // If this cell isn't for the cohort in this rule, skip
            if (cellValue !== cohortName) {
              return;
            }

            // Get current timeslot from header
            const headerRow = scheduleRange.values[0];
            const timeslot = headerRow[col];

            // Check if this timeslot is blacklisted for this cohort
            for (let i = 1; i < blacklistConfig.length; i++) {
              for (let j = 0; j < blacklistConfig[i].length; j++) {
                if (blacklistConfig[i][j] === timeslot) {
                  brokenRules.push(
                    `The cohort '${cohortName}' is not allowed to have any class during '${timeslot}' as defined in the CohortBlacklist rule. See column '${cohortBlacklistConfig.column}' on the Rules sheet.`
                  );
                  break;
                }
              }
            }
          });

          // Check OneClassAtATime rule
          if (oneClassAtATimeEnabled) {
            // Get current timeslot from header
            const headerRow = scheduleRange.values[0];
            const timeslot = headerRow[col];
            const cohort = cellValue;

            // Check if this cohort is taking any other class in this timeslot
            for (let otherRow = 1; otherRow < scheduleRange.rowCount; otherRow++) {
              // Skip the current row (same class)
              if (otherRow === row) {
                continue;
              }

              const otherClassName = scheduleRange.values[otherRow][0];
              const otherCellValue = scheduleRange.values[otherRow][col];

              // If this cohort is assigned to another class in the same timeslot
              if (otherCellValue === cohort) {
                brokenRules.push(
                  `The cohort '${cohort}' is scheduled for both '${className}' and '${otherClassName}' during '${timeslot}'. Cohorts can only attend one class at a time according to the OneClassAtATime rule.`
                );
                break; // One conflict is enough to report the issue
              }
            }
          }

          // add errors
          if (brokenRules.length != 0) {
            console.log(`adding broken rules ${brokenRules}`);
            const cell = scheduleRange.getCell(row, col);
            cell.format.fill.color = "red";

            // Create main comment
            const mainComment = "Rule Infractions:";
            try {
              const comment = scheduleSheet.comments.add(cell, mainComment);

              // Add each broken rule as a reply
              brokenRules.forEach((rule) => {
                comment.replies.add(rule);
              });

              console.log("excel-scheduler added comment with infractions successfully");
            } catch (error) {
              console.error("excel-scheduler error adding comment:", error);
            }
          }
        }
      }

      await context.sync();
    });
  } catch (error) {
    console.error("excel-scheduler error in validation script:", error);
  }

  // Calculate and log execution time
  const endTime = performance.now();
  const executionTime = (endTime - startTime).toFixed(2);
  console.log(`excel-scheduler run function complete, execution time: ${executionTime} ms`);
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
    const rulesSheet = context.workbook.worksheets.getItem("Rules");
    const validationRange = rulesSheet.getCell(0, columnIndex);

    // Add data validation
    validationRange.dataValidation.clear();

    validationRange.dataValidation.prompt = {
      message: validationMessage,
      showPrompt: true,
      title: validationTitle,
    };
  } catch (error) {
    console.error("Error: ", error);
  }
}

export async function clear() {
  // Start timer
  const clearStartTime = performance.now();

  try {
    await Excel.run(async (context) => {
      console.log("excel-scheduler clearing formatting");

      const sheet = context.workbook.worksheets.getItem("Schedule");
      const entireUsedRange = sheet.getUsedRange();
      entireUsedRange.format.fill.clear();

      // Clear all comments from the sheet
      sheet.load(["comments"]);
      await context.sync();

      sheet.comments.items.forEach((comment) => {
        comment.delete();
      });

      await context.sync();
    });

    // Calculate and log execution time
    const clearEndTime = performance.now();
    const clearExecutionTime = (clearEndTime - clearStartTime).toFixed(2);
    console.log(`excel-scheduler clear function execution time: ${clearExecutionTime} ms`);
  } catch (error) {
    console.error("excel-scheduler error in clear script:", error);
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
