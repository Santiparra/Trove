import type { Book } from "../../bindings/github.com/Santiparra/Trove/internal/books/models";
import type { Choice } from "../../bindings/github.com/Santiparra/Trove/internal/desktop/models";

export default function StatusSelect({
  book,
  options,
  disabled,
  onChange,
}: {
  book: Book;
  options: Choice[];
  disabled: boolean;
  onChange: (book: Book, status: string) => void;
}) {
  return (
    <select
      className={`quick-status status-${book.cycle.status}`}
      aria-label={`Estado de ${book.title}`}
      value={book.cycle.status}
      disabled={disabled}
      onChange={(event) => onChange(book, event.target.value)}
    >
      {options.map((option) => (
        <option key={option.value} value={option.value}>
          {option.label}
        </option>
      ))}
    </select>
  );
}
