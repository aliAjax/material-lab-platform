import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";
import { LoginPage } from "./LoginPage";
import { AuthProvider } from "../../state/AuthContext";

describe("login", () => {
  it("validates credentials before issuing a request", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
    render(
      <QueryClientProvider client={new QueryClient()}>
        <MemoryRouter>
          <AuthProvider>
            <LoginPage />
          </AuthProvider>
        </MemoryRouter>
      </QueryClientProvider>,
    );
    fireEvent.change(screen.getByLabelText("用户名 *"), {
      target: { value: "tester" },
    });
    fireEvent.change(screen.getByLabelText("密码 *"), {
      target: { value: "short" },
    });
    fireEvent.click(screen.getByRole("button", { name: "登录" }));
    await waitFor(() =>
      expect(screen.getByText("密码至少 8 位")).toBeInTheDocument(),
    );
    expect(fetchMock).not.toHaveBeenCalled();
  });
});
