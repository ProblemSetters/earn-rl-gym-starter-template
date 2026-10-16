// Deletion in the real UI. Prints {scenario: [passed, detail]} as JSON.
const { chromium } = require("playwright");

const results = {};
async function scenario(name, fn) {
  try {
    results[name] = [true, (await fn()) || "ok"];
  } catch (error) {
    results[name] = [false, `${error.name}: ${error.message}`.split("\n")[0]];
  }
}

async function main() {
  const baseURL = process.argv[2];
  const browser = await chromium.launch({ headless: true, args: ["--no-sandbox"] });
  try {
    const item = (page, title) => page.getByRole("listitem").filter({ hasText: title });

    // 1. Delete removes the item from the list.
    await scenario("removed", async () => {
      const page = await browser.newPage();
      await page.goto(baseURL, { waitUntil: "networkidle" });
      const button = item(page, "Plan sprint review").getByRole("button", { name: "Delete" });
      await button.waitFor({ state: "visible", timeout: 5000 });
      await button.click();
      await page.getByText("Plan sprint review", { exact: true }).waitFor({ state: "detached", timeout: 5000 });
      await page.close();
      return "the deleted task left the list";
    });

    // 2. A fresh page load still shows it gone, and the other task kept.
    await scenario("reload", async () => {
      const page = await browser.newPage();
      await page.goto(baseURL, { waitUntil: "networkidle" });
      await item(page, "Update onboarding guide").waitFor({ state: "visible", timeout: 5000 });
      if (await page.getByText("Plan sprint review", { exact: true }).count()) throw new Error("the deleted task is back after a reload");
      await page.close();
      return "still gone after a reload";
    });

    // 3. When the DELETE request fails, the task stays visible and an error is shown.
    await scenario("failure", async () => {
      const page = await browser.newPage();
      await page.route("**/api/todos/2", (route) =>
        route.request().method() === "DELETE"
          ? route.fulfill({ status: 500, contentType: "application/json", body: JSON.stringify({ error: "Could not delete the task" }) })
          : route.continue());
      await page.goto(baseURL, { waitUntil: "networkidle" });
      const row = item(page, "Update onboarding guide");
      await row.getByRole("button", { name: "Delete" }).click();
      await page.getByRole("alert").filter({ hasText: /\S/ }).waitFor({ state: "visible", timeout: 5000 });
      await row.waitFor({ state: "visible", timeout: 1000 });
      await page.close();
      return "the task stayed and an error was shown";
    });
  } finally {
    await browser.close();
  }
  process.stdout.write(JSON.stringify(results) + "\n");
}

main().catch((error) => {
  process.stderr.write(`${error.name}: ${error.message}\n`);
  process.exitCode = 1;
});
