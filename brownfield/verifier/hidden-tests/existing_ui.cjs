// The list and add flow in the real UI. Prints {scenario: [passed, detail]} as JSON.
const { chromium } = require("playwright");

async function main() {
  const baseURL = process.argv[2];
  const browser = await chromium.launch({ headless: true, args: ["--no-sandbox"] });
  const results = {};
  try {
    const page = await browser.newPage();
    await page.goto(baseURL, { waitUntil: "networkidle" });
    try {
      for (const title of ["Plan sprint review", "Update onboarding guide"]) {
        await page.getByRole("listitem").filter({ hasText: title }).waitFor({ state: "visible", timeout: 5000 });
      }
      const input = page.getByLabel("What needs to get done?");
      await input.fill("Ship release");
      await page.getByRole("button", { name: "Add task" }).click();
      await page.getByRole("listitem").filter({ hasText: "Ship release" }).waitFor({ state: "visible", timeout: 5000 });
      if ((await input.inputValue()) !== "") throw new Error("the form was not cleared after adding");
      results.listAdd = [true, "listed the tasks and added one"];
    } catch (error) {
      results.listAdd = [false, `${error.name}: ${error.message}`.split("\n")[0]];
    }
  } finally {
    await browser.close();
  }
  process.stdout.write(JSON.stringify(results) + "\n");
}

main().catch((error) => {
  process.stderr.write(`${error.name}: ${error.message}\n`);
  process.exitCode = 1;
});
