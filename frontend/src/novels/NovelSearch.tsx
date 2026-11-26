import { useEffect, useRef, useState } from "react";
import * as Novels from "../../bindings/github.com/Santiparra/Trove/internal/desktop/novels";
import type { Novel } from "../../bindings/github.com/Santiparra/Trove/internal/novels/models";
import type {
  Candidate,
  SearchResult,
} from "../../bindings/github.com/Santiparra/Trove/internal/novelcatalog/models";
import { errorMessage } from "../books/presentation";
import Icon from "../components/Icon";
import { representedNovel } from "./presentation";

export default function NovelSearch({
  value,
  onChange,
  disabled,
  localOnly,
  authorID = "",
  onNovel,
  onCandidate,
  onAll,
  catalogOnly = false,
}: {
  value: string;
  onChange: (value: string) => void;
  disabled: boolean;
  localOnly: boolean;
  authorID?: string;
  onNovel: (novel: Novel) => void;
  onCandidate: (candidate: Candidate) => void;
  onAll: (query: string) => void;
  catalogOnly?: boolean;
}) {
  const [open, setOpen] = useState(false);
  const [locals, setLocals] = useState<Novel[] | null>(null);
  const [remote, setRemote] = useState<SearchResult | null>(null);
  const [error, setError] = useState("");
  const [remoteError, setRemoteError] = useState("");
  const [searching, setSearching] = useState(false);
  const root = useRef<HTMLDivElement>(null);
  useEffect(() => {
    let active = true;
    setLocals(null);
    setError("");
    if (!open || !value.trim()) return;
    const timer = setTimeout(() => {
      (catalogOnly ? Promise.resolve([]) : Novels.List(value.trim(), ""))
        .then((items) => {
          if (active) setLocals(items);
        })
        .catch((reason) => {
          if (active) setError(errorMessage(reason));
        });
    }, 140);
    return () => {
      active = false;
      clearTimeout(timer);
    };
  }, [open, value, catalogOnly]);
  useEffect(() => {
    let active = true;
    setRemote(null);
    setRemoteError("");
    setSearching(false);
    if (!open || !locals || localOnly || value.trim().length < 2) return;
    setSearching(true);
    const timer = setTimeout(() => {
      Novels.SearchCatalog(value.trim(), authorID, 1, "popover")
        .then((items) => {
          if (active) setRemote(items);
        })
        .catch((reason) => {
          if (active) setRemoteError(errorMessage(reason));
        })
        .finally(() => {
          if (active) setSearching(false);
        });
    }, 350);
    return () => {
      active = false;
      clearTimeout(timer);
    };
  }, [open, value, locals, localOnly, authorID]);
  useEffect(() => {
    function outside(event: PointerEvent) {
      if (!root.current?.contains(event.target as Node)) setOpen(false);
    }
    document.addEventListener("pointerdown", outside);
    return () => document.removeEventListener("pointerdown", outside);
  }, []);
  const candidates =
    remote?.items.filter((item) => !representedNovel(item, locals || [])) || [];
  function all() {
    setOpen(false);
    onAll(value.trim());
  }
  return (
    <div className="search-root search-box" ref={root}>
      <Icon name="search" size={17} />
      <input
        type="search"
        aria-label="Buscar novelas"
        value={value}
        placeholder="Título, autoría o ISBN…"
        disabled={disabled}
        onChange={(event) => {
          onChange(event.target.value);
          setOpen(true);
        }}
        onFocus={() => setOpen(true)}
        aria-expanded={open && !!value.trim()}
        onKeyDown={(event) => {
          if (event.key === "Escape") setOpen(false);
          if (event.key === "Enter") {
            event.preventDefault();
            if (catalogOnly) setOpen(true);
            else all();
          }
        }}
      />
      {open && !!value.trim() && (
        <div
          className="search-popover"
          aria-label="Resultados de búsqueda de novelas"
        >
          {error && (
            <p className="search-feedback" role="alert">
              {error}
            </p>
          )}
          {locals === null && !error && (
            <p className="search-feedback">Buscando en tu biblioteca…</p>
          )}
          {!!locals?.length && (
            <>
              <p className="search-group-label">Tu biblioteca</p>
              <ul>
                {locals.slice(0, 5).map((n) => (
                  <li key={n.id}>
                    <button
                      disabled={disabled}
                      onClick={() => {
                        setOpen(false);
                        onNovel(n);
                      }}
                    >
                      <Icon name="books" size={18} />
                      <span>
                        <strong>{n.title}</strong>
                        <span>{n.authors.join(" · ")}</span>
                      </span>
                      <Icon name="chevron" size={15} />
                    </button>
                  </li>
                ))}
              </ul>
            </>
          )}
          {locals?.length === 0 && !catalogOnly && (
            <p className="search-feedback">Sin coincidencias locales.</p>
          )}
          {localOnly ? (
            <p className="search-feedback">
              Modo solo local · Proveedores desactivados
            </p>
          ) : searching ? (
            <p className="search-feedback" role="status">
              Buscando en RanobeDB…
            </p>
          ) : null}
          {remoteError && (
            <p className="search-feedback" role="alert">
              {remoteError}
            </p>
          )}
          {!!candidates.length && (
            <>
              <p className="search-group-label">RanobeDB</p>
              <ul>
                {candidates.slice(0, 6).map((item) => (
                  <li key={item.id}>
                    <button
                      disabled={disabled}
                      onClick={() => {
                        setOpen(false);
                        onCandidate(item);
                      }}
                    >
                      <Icon name="books" size={18} />
                      <span>
                        <strong>{item.title}</strong>
                        <span>{item.originalTitle}</span>
                      </span>
                      <Icon name="chevron" size={15} />
                    </button>
                  </li>
                ))}
              </ul>
            </>
          )}
          {!catalogOnly && (
            <button className="search-all" disabled={disabled} onClick={all}>
              Ver todos los resultados <Icon name="chevron" size={15} />
            </button>
          )}
        </div>
      )}
    </div>
  );
}
