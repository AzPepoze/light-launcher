const url = process.argv[2];
const timeoutMs = 30_000;
const intervalMs = 300;

if (!url) {
  console.error("Usage: bun scripts/wait-for-url.ts <url>");
  process.exit(1);
}

async function hasResponse(): Promise<boolean> {
  try {
    await fetch(url, { signal: AbortSignal.timeout(intervalMs) });
    return true;
  } catch {
    return false;
  }
}

const deadline = Date.now() + timeoutMs;

while (Date.now() < deadline) {
  if (await hasResponse()) {
    process.exit(0);
  }
  await Bun.sleep(intervalMs);
}

console.error(`Timed out after 30 seconds waiting for ${url}`);
process.exit(1);
