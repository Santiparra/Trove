import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { act, cleanup, render } from "@testing-library/react";
import BookSearch from "./BookSearch";

const api = vi.hoisted(() => ({ local: vi.fn(), remote: vi.fn() }));
vi.mock(
  "../../bindings/github.com/Santiparra/Trove/internal/desktop/library",
  () => ({ SearchBooks: api.local }),
);
vi.mock(
  "../../bindings/github.com/Santiparra/Trove/internal/desktop/catalog",
  () => ({ Search: api.remote }),
);

beforeEach(() => {
  vi.useFakeTimers();
  api.local.mockReset().mockResolvedValue({ matches: [], total: 0 });
  api.remote.mockReset().mockResolvedValue({ works: [], warnings: [] });
});
afterEach(() => {
  cleanup();
  vi.useRealTimers();
});

it("espera la pausa de escritura, cancela consultas pendientes y busca local antes del proveedor", async () => {
  const props = {
    medium: "books",
    disabled: false,
    configured: true,
    onChange: vi.fn(),
    onBook: vi.fn(),
    onAll: vi.fn(),
  };
  const page = render(<BookSearch {...props} value="" />);
  const input = page.getByRole("searchbox");
  act(() => {
    input.focus();
  });
  page.rerender(<BookSearch {...props} value="A" />);
  await act(() => vi.advanceTimersByTimeAsync(100));
  page.rerender(<BookSearch {...props} value="Au" />);
  await act(() => vi.advanceTimersByTimeAsync(139));
  expect(api.local).not.toHaveBeenCalled();
  expect(api.remote).not.toHaveBeenCalled();
  await act(() => vi.advanceTimersByTimeAsync(1));
  expect(api.local).toHaveBeenCalledExactlyOnceWith("Au", "books");
  await act(() => vi.advanceTimersByTimeAsync(349));
  expect(api.remote).not.toHaveBeenCalled();
  page.rerender(<BookSearch {...props} value="Aut" />);
  await act(() => vi.advanceTimersByTimeAsync(140));
  expect(api.local).toHaveBeenLastCalledWith("Aut", "books");
  await act(() => vi.advanceTimersByTimeAsync(350));
  expect(api.remote).toHaveBeenCalledExactlyOnceWith("Aut");
});
it("en solo local busca en la biblioteca sin programar una petición remota", async () => {
  render(
    <BookSearch
      medium="books"
      value="Sintético"
      disabled={false}
      configured
      localOnly
      onChange={vi.fn()}
      onBook={vi.fn()}
      onAll={vi.fn()}
    />,
  );
  act(() => {
    document.querySelector<HTMLInputElement>("input")!.focus();
  });
  await act(() => vi.advanceTimersByTimeAsync(1000));
  expect(api.local).toHaveBeenCalledExactlyOnceWith("Sintético", "books");
  expect(api.remote).not.toHaveBeenCalled();
});
