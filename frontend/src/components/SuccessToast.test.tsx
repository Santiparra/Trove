import { afterEach, expect, it, vi } from "vitest";
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
} from "@testing-library/react";
import SuccessToast from "./SuccessToast";
afterEach(() => {
  cleanup();
  vi.useRealTimers();
});
it("desaparece solo y pausa el temporizador mientras se lee con el mouse", async () => {
  vi.useFakeTimers();
  const dismiss = vi.fn();
  render(<SuccessToast message="Guardado" onDismiss={dismiss} />);
  await act(() => vi.advanceTimersByTimeAsync(4000));
  fireEvent.pointerEnter(screen.getByRole("status"));
  await act(() => vi.advanceTimersByTimeAsync(5000));
  expect(dismiss).not.toHaveBeenCalled();
  fireEvent.pointerLeave(screen.getByRole("status"));
  await act(() => vi.advanceTimersByTimeAsync(5000));
  expect(dismiss).toHaveBeenCalledOnce();
});
