import assert from "node:assert/strict";
import { after, before, test } from "node:test";
import { chromium } from "playwright";
import { createServer } from "vite";

let server;
let browser;
let baseURL;
before(async () => {
  server = await createServer({ server: { host: "127.0.0.1", port: 0 } });
  await server.listen();
  baseURL = server.resolvedUrls.local[0];
  browser = await chromium.launch();
});
after(async () => {
  await browser?.close();
  await server?.close();
});

async function fixture(t, viewport = { width: 1280, height: 720 }) {
  const page = await browser.newPage({ viewport });
  page.setDefaultTimeout(5000);
  t.after(() => page.close());
  const errors = [];
  page.on("pageerror", (error) => errors.push(error.message));
  t.after(() => assert.deepEqual(errors, []));
  await page.route(
    (url) => url.pathname.startsWith("/api/"),
    (route) => {
      const path = new URL(route.request().url()).pathname;
      const data =
        path === "/api/me"
          ? { id: 1, username: "test", admin: true }
          : path.endsWith("/reviews")
            ? []
            : {};
      return route.fulfill({ json: data });
    },
  );
  await page.goto(`${baseURL}tests/browser/fixtures/dialog-scroll.html`);
  await page.getByRole("button", { name: "Open review", exact: true }).waitFor();
  return page;
}

async function locked(page, expected) {
  await page.waitForFunction(
    (value) => (getComputedStyle(document.body).overflowY === "hidden") === value,
    expected,
  );
}

async function scrollable(page) {
  await locked(page, false);
  assert.equal(await page.locator("body").getAttribute("data-scroll-locked"), null);
  assert.equal(await page.evaluate(() => document.body.style.overflow), "");
  await page.mouse.move(1100, 650);
  const before = await page.evaluate(() => scrollY);
  await page.mouse.wheel(0, 300);
  await page.waitForFunction((y) => scrollY > y, before);
}

for (const viewport of [
  { width: 1280, height: 720 },
  { width: 390, height: 844 },
]) {
  test(`review close restores scrolling (${viewport.width}px)`, async (t) => {
    const page = await fixture(t, viewport);
    for (const close of ["button", "escape", "overlay"]) {
      await page.evaluate(() => scrollTo(0, 0));
      await page.getByRole("button", { name: "Open review", exact: true }).click();
      await page.getByRole("textbox", { name: "评价内容" }).waitFor();
      await locked(page, true);
      if (close === "button")
        await page.getByRole("button", { name: "Close", exact: true }).click();
      else if (close === "escape") await page.keyboard.press("Escape");
      else
        await page
          .locator('[data-slot="dialog-overlay"]')
          .click({ position: { x: 2, y: 2 } });
      await page.getByRole("dialog").waitFor({ state: "detached" });
      // Put the wheel inside the viewport on both desktop and narrow screens.
      await page.mouse.move(100, 600);
      await locked(page, false);
      await page.mouse.wheel(0, 300);
      await page.waitForFunction(() => scrollY > 0);
      assert.equal(await page.evaluate(() => document.body.style.overflow), "");
    }
  });
}

test("nested attachment only unlocks after the review closes", async (t) => {
  const page = await fixture(t);
  await page.getByRole("button", { name: "Open review", exact: true }).click();
  const review = page.getByRole("dialog");
  await page.getByRole("textbox", { name: "评价内容" }).waitFor();
  await review.locator("summary").click();
  await review.getByRole("button", { name: /preview.png/ }).click();
  await page.getByRole("dialog", { name: "preview.png" }).waitFor();
  await page.keyboard.press("Escape");
  await page.getByRole("dialog", { name: "preview.png" }).waitFor({ state: "detached" });
  await locked(page, true);
  await page.keyboard.press("Escape");
  await review.waitFor({ state: "detached" });
  await scrollable(page);
});

test("plain confirmation and attachment dialogs restore scrolling", async (t) => {
  const page = await fixture(t);
  for (const name of ["Open confirm", /preview.png/]) {
    await page.evaluate(() => scrollTo(0, 0));
    await page.getByRole("button", { name }).click();
    await page.getByRole("dialog").waitFor();
    await locked(page, true);
    await page.keyboard.press("Escape");
    await page.getByRole("dialog").waitFor({ state: "detached" });
    await scrollable(page);
  }
});

test("editor fullscreen exit and unmount release their own lock", async (t) => {
  const page = await fixture(t);
  await page.getByRole("button", { name: "Toggle editor" }).click();
  const fullscreen = page.getByRole("button", {
    name: "切换全屏(ctrl + 0)",
    exact: true,
  });
  await fullscreen.click();
  await locked(page, true);
  await fullscreen.click();
  await scrollable(page);
  await fullscreen.click();
  await locked(page, true);
  // Simulate navigation unmounting a fullscreen editor.
  await page
    .getByRole("button", { name: "Toggle editor" })
    .evaluate((button) => button.click());
  await scrollable(page);
});

test("fullscreen editor preserves the enclosing review lock", async (t) => {
  const page = await fixture(t);
  await page.getByRole("button", { name: "Open review", exact: true }).click();
  await page.getByRole("textbox", { name: "评价内容" }).waitFor();
  const fullscreen = page.getByRole("button", {
    name: "切换全屏(ctrl + 0)",
    exact: true,
  });
  await fullscreen.click();
  await locked(page, true);
  await fullscreen.click();
  await locked(page, true);
  await fullscreen.click();
  await page.keyboard.press("Escape");
  await page.getByRole("dialog").waitFor({ state: "detached" });
  await scrollable(page);
});
