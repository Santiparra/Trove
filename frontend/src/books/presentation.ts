import type { Book } from "../../bindings/github.com/Santiparra/Trove/internal/books/models";
import type { Choice } from "../../bindings/github.com/Santiparra/Trove/internal/desktop/models";

export function labelFor(options: Choice[], value: string) {
  return options.find((item) => item.value === value)?.label ?? value;
}
export function dateLabel(value: string) {
  if (!value) return "Sin registrar";
  const [year, month, day] = value.split("-");
  return `${day}/${month}/${year}`;
}
export function seriesLabel(book: Book) {
  return book.series
    ? `${book.series.name}${book.series.position !== null ? ` · #${book.series.position}` : ""}`
    : "Sin serie";
}
export function errorMessage(error: unknown) {
  return error instanceof Error ? error.message : String(error);
}
