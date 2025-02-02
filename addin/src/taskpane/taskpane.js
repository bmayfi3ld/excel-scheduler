/*
 * Copyright (c) Microsoft Corporation. All rights reserved. Licensed under the MIT license.
 * See LICENSE in the project root for license information.
 */

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

      // Get the used range of the sheet
      const scheduleRange = scheduleSheet.getUsedRange();
      scheduleRange.load(["values", "rowCount", "columnCount"]);
      await context.sync();

      console.log(`Sheet dimensions: ${scheduleRange.rowCount} rows, ${scheduleRange.columnCount} columns`);
      console.log("Full range values:", scheduleRange.values);

      // Get Values for Rules
      const rulesSheet = context.workbook.worksheets.getItem("Rules");

      const allClasses = ["PKA", "2nd"];
      console.log("Valid values:", allClasses);

      // Iterate through each cell, starting from row 2 (skip header) and column 2 (skip first column)
      for (let row = 1; row < scheduleRange.rowCount; row++) {
        for (let col = 1; col < scheduleRange.columnCount; col++) {
          const cellValue = scheduleRange.values[row][col];
          console.log(`Checking cell at row ${row + 1}, column ${col + 1}:`, {
            value: cellValue,
            isEmpty: !cellValue,
            isValid: allClasses.includes(cellValue),
          });

          // Skip empty cells
          if (!cellValue) {
            console.log("Skipping empty cell");
            continue;
          }

          // Check if the value is not in the valid list
          if (!allClasses.includes(cellValue)) {
            console.log(`Invalid value found: "${cellValue}"`);

            // Get the specific cell
            const cell = scheduleRange.getCell(row, col);
            console.log("Got cell reference");

            // Set fill color to red
            cell.format.fill.color = "red";
            console.log("Set cell color to red");

            try {
              // Add comment using Excel.Comment
              scheduleSheet.comments.add(cell, "This class isn't in the total list of classes");
              console.log("Added comment successfully");
            } catch (commentError) {
              console.error("Error adding comment:", commentError);
            }
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
