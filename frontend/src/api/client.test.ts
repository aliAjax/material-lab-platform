import { afterEach, describe, expect, it, vi } from "vitest";
import { api, authToken, queryString, writeOptions } from "./client";

describe("API client", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    authToken.set("");
  });
  it("builds stable filter query without empty values", () =>
    expect(queryString({ status: "pending", cursor: undefined, q: "" })).toBe(
      "?status=pending",
    ));
  it("adds request identity, auth and idempotency headers", async () => {
    authToken.set("access-token");
    const fetchMock = vi
      .fn()
      .mockResolvedValue(
        new Response(JSON.stringify({ id: "1" }), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        }),
      );
    vi.stubGlobal("fetch", fetchMock);
    await api("/records", writeOptions("POST", { name: "sample" }));
    const [, init] = fetchMock.mock.calls[0];
    expect((init.headers as Headers).get("Authorization")).toBe(
      "Bearer access-token",
    );
    expect((init.headers as Headers).get("Idempotency-Key")).toBeTruthy();
    expect((init.headers as Headers).get("X-Request-ID")).toBeTruthy();
  });
  it("surfaces structured validation errors", async () => {
    vi.stubGlobal(
      "fetch",
      vi
        .fn()
        .mockResolvedValue(
          new Response(
            JSON.stringify({
              message: "校验失败",
              fields: { sample_no: "不能为空" },
              request_id: "req-1",
            }),
            { status: 422, headers: { "Content-Type": "application/json" } },
          ),
        ),
    );
    await expect(api("/samples")).rejects.toMatchObject({
      status: 422,
      problem: { message: "校验失败", request_id: "req-1", fields: { sample_no: "不能为空" } },
    });
  });
});
