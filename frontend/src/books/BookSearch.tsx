import { useEffect, useId, useRef, useState } from "react";
import type {
  Book,
  SearchResults,
} from "../../bindings/github.com/Santiparra/Trove/internal/books/models";
import * as Library from "../../bindings/github.com/Santiparra/Trove/internal/desktop/library";
import Icon from "../components/Icon";
import {
  errorMessage,
  seriesLabel,
  representedInLibrary,
} from "./presentation";
import * as Catalog from "../../bindings/github.com/Santiparra/Trove/internal/desktop/catalog";
import type { Work } from "../../bindings/github.com/Santiparra/Trove/internal/catalog/models";

export default function BookSearch({
  medium,
  value,
  disabled,
  onChange,
  onBook,
  onAll,
  onWork,
  configured = false,
  catalogOnly = false,
  localOnly = false,
}: {
  medium: string;
  value: string;
  disabled: boolean;
  onChange: (query: string) => void;
  onBook: (book: Book) => void;
  onAll: (query: string) => void;
  onWork?: (work: Work) => void;
  configured?: boolean;
  catalogOnly?: boolean;
  localOnly?: boolean;
}) {
  const [open, setOpen] = useState(false);
  const [results, setResults] = useState<SearchResults | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [works, setWorks] = useState<Work[]>([]);
  const [remoteLoading, setRemoteLoading] = useState(false);
  const [remoteError, setRemoteError] = useState("");
  const root = useRef<HTMLDivElement>(null);
  const panelID = useId();
  useEffect(() => {
    if (!open || !value.trim()) {
      setResults(null);
      setLoading(false);
      return;
    }
    let active = true;
    setLoading(true);
    setError("");
    setResults(null);
    const timer = setTimeout(() => {
      (catalogOnly
        ? Promise.resolve({ matches: [], total: 0 })
        : Library.SearchBooks(value.trim(), medium)
      )
        .then((items) => {
          if (active) setResults(items);
        })
        .catch((reason) => {
          if (active) setError(errorMessage(reason));
        })
        .finally(() => {
          if (active) setLoading(false);
        });
    }, 140);
    return () => {
      active = false;
      clearTimeout(timer);
    };
  }, [open, value, medium, catalogOnly]);
  useEffect(() => {
    setWorks([]);
    setRemoteError("");
    if (
      !open ||
      !results ||
      medium !== "books" ||
      !configured ||
      localOnly ||
      value.trim().length < 2
    ) {
      setRemoteLoading(false);
      return;
    }
    let active = true;
    setRemoteLoading(true);
    const timer = setTimeout(() => {
      Catalog.Search(value.trim())
        .then((result) => {
          if (active) setWorks(result.works || []);
        })
        .catch((reason) => {
          if (active) setRemoteError(errorMessage(reason));
        })
        .finally(() => {
          if (active) setRemoteLoading(false);
        });
    }, 350);
    return () => {
      active = false;
      clearTimeout(timer);
    };
  }, [open, results, medium, configured, value, localOnly]);
  useEffect(() => {
    function outside(event: PointerEvent) {
      if (!root.current?.contains(event.target as Node)) setOpen(false);
    }
    document.addEventListener("pointerdown", outside);
    return () => document.removeEventListener("pointerdown", outside);
  }, []);
  function all() {
    setOpen(false);
    onAll(value.trim());
  }
  return (
    <div
      className="search-root"
      ref={root}
      onBlur={(event) => {
        if (!event.currentTarget.contains(event.relatedTarget as Node | null))
          setOpen(false);
      }}
    >
      <label className="search-box">
        <Icon name="search" size={18} />
        <input
          type="search"
          value={value}
          disabled={disabled}
          aria-label="Buscar por título"
          aria-controls={panelID}
          aria-expanded={open && !!value.trim()}
          placeholder={
            medium === "books"
              ? "Título, autor, serie o ISBN…"
              : "Buscar por título en tu biblioteca…"
          }
          onFocus={() => setOpen(true)}
          onChange={(event) => {
            onChange(event.target.value);
            setOpen(true);
          }}
          onKeyDown={(event) => {
            if (event.key === "Escape") {
              event.preventDefault();
              setOpen(false);
            }
            if (event.key === "Enter" && value.trim()) {
              event.preventDefault();
              if (catalogOnly) setOpen(true);
              else all();
            }
            if (event.key === "ArrowDown") {
              event.preventDefault();
              setOpen(true);
              root.current
                ?.querySelector<HTMLButtonElement>(".search-popover button")
                ?.focus();
            }
          }}
        />
      </label>
      {open && value.trim() && (
        <div
          className="search-popover"
          id={panelID}
          role="region"
          aria-label="Resultados de búsqueda"
          onKeyDown={(event) => {
            if (event.key === "Escape") {
              root.current?.querySelector("input")?.focus();
              setOpen(false);
            }
          }}
        >
          {loading && (
            <p className="search-feedback" role="status">
              Buscando en tu biblioteca…
            </p>
          )}
          {error && (
            <p className="search-feedback" role="alert">
              {error}
            </p>
          )}
          {results && (
            <>
              <ul>
                {results.matches.map((book) => (
                  <li key={book.id}>
                    <button
                      disabled={disabled}
                      onClick={() => {
                        setOpen(false);
                        onBook(book);
                      }}
                      aria-label={`Abrir resultado ${book.title}`}
                    >
                      <Icon name="books" size={18} />
                      <span>
                        {book.series && <small>{seriesLabel(book)}</small>}
                        <strong>{book.title}</strong>
                        <span>
                          {book.authors.join(" · ") ||
                            "Autoría sin especificar"}
                        </span>
                      </span>
                      <Icon name="chevron" size={15} />
                    </button>
                  </li>
                ))}
              </ul>
              {!results.total && !catalogOnly && (
                <p className="search-feedback">
                  No hay coincidencias en la biblioteca local.
                </p>
              )}
              {remoteLoading && (
                <p className="search-feedback" role="status">
                  Buscando en Hardcover…
                </p>
              )}
              {remoteError && (
                <p className="search-feedback" role="alert">
                  {remoteError}
                </p>
              )}
              {medium === "books" && localOnly && (
                <p className="search-feedback">
                  Modo solo local · Proveedores desactivados
                </p>
              )}
              {medium === "books" && !configured && !localOnly && (
                <p className="search-feedback">
                  Catálogo sin configurar · Claves en el encabezado
                </p>
              )}
              {!!works.length && (
                <>
                  <p className="search-group-label">Hardcover</p>
                  <ul>
                    {works
                      .filter(
                        (work) => !representedInLibrary(work, results.matches),
                      )
                      .slice(0, 6)
                      .map((work) => (
                        <li key={work.id}>
                          <button
                            disabled={disabled}
                            onClick={() => {
                              setOpen(false);
                              onWork?.(work);
                            }}
                          >
                            <Icon name="books" size={18} />
                            <span>
                              <strong>{work.title}</strong>
                              <span>
                                {work.authors?.join(" · ") ||
                                  "Autoría sin especificar"}
                              </span>
                            </span>
                            <Icon name="chevron" size={15} />
                          </button>
                        </li>
                      ))}
                  </ul>
                </>
              )}
              {!catalogOnly && (
                <button
                  className="search-all"
                  onClick={all}
                  disabled={disabled}
                >
                  Ver todos los resultados (
                  {results.total +
                    works.filter(
                      (work) => !representedInLibrary(work, results.matches),
                    ).length}
                  ) <Icon name="chevron" size={15} />
                </button>
              )}
            </>
          )}
        </div>
      )}
    </div>
  );
}
