// Simulated latency for the mock API. Kept tiny so the static app feels
// instant while still showing loading skeletons briefly. The per-call argument
// is capped by MOCK_DELAY_MS, so a stale `delay(600)` can never re-introduce
// noticeable lag. Set MOCK_DELAY_MS to 0 for truly instant responses.
export const MOCK_DELAY_MS = 120;

export const delay = (ms = MOCK_DELAY_MS) =>
  new Promise((r) => setTimeout(r, Math.min(ms, MOCK_DELAY_MS)));
