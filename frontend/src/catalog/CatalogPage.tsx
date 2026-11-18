import { useState } from "react";
import * as Catalog from "../../bindings/github.com/Santiparra/Trove/internal/desktop/catalog";
import type {
  CatalogPreview,
  Info,
} from "../../bindings/github.com/Santiparra/Trove/internal/desktop/models";
import type { Work } from "../../bindings/github.com/Santiparra/Trove/internal/catalog/models";
import type { Book } from "../../bindings/github.com/Santiparra/Trove/internal/books/models";
import BookProfile from "../books/BookProfile";
import { errorMessage } from "../books/presentation";
import EditionPicker from "./EditionPicker";

export default function CatalogPage({
  initial,
  target,
  info,
  configured,
  onBack,
  onDone,
  onSearch,
  onBusy,
}: {
  initial: CatalogPreview;
  target: Book | null;
  info: Info;
  configured: boolean;
  onBack: () => void;
  onDone: (book: Book, message: string) => void;
  onSearch: (query: string) => void;
  onBusy: (value: boolean) => void;
}) {
  const [preview, setPreview] = useState(initial);
  const [series, setSeries] = useState(
    initial.snapshot.work.series?.length === 1
      ? initial.snapshot.work.series[0].id
      : "",
  );
  const [status, setStatus] = useState("to-read");
  const [picking, setPicking] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const snapshot = preview.snapshot;
  const match = preview.match;
  const local = target || match?.book || null;
  async function action(operation: () => PromiseLike<unknown>) {
    if (busy) return;
    setBusy(true);
    onBusy(true);
    setError("");
    try {
      await operation();
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setBusy(false);
      onBusy(false);
    }
  }
  function accept(value: CatalogPreview) {
    if (value.match?.kind === "existing") {
      onDone(value.match.book, "");
      return;
    }
    setPreview(value);
    setSeries(
      value.snapshot.work.series?.length === 1
        ? value.snapshot.work.series[0].id
        : "",
    );
    setPicking(false);
  }
  function openWork(work: Work) {
    void action(async () =>
      accept(
        await (work.matchedEditionId
          ? Catalog.Preview(work.matchedEditionId)
          : Catalog.OpenWork(work.id)),
      ),
    );
  }
  function add(selectedStatus: string) {
    setStatus(selectedStatus);
    void action(async () => {
      const result = await Catalog.Add(preview.token, series, selectedStatus);
      if (result.kind === "review") setPreview({ ...preview, match: result });
      else onDone(result.book, result.message);
    });
  }
  return (
    <BookProfile
      snapshot={snapshot}
      busy={busy}
      onBack={onBack}
      onSearch={onSearch}
      management={
        <>
          {local ? (
            <>
              <p className="local-match">
                {match?.kind === "existing"
                  ? "En tu biblioteca"
                  : "Coincidencia local"}
                : <strong>{local.title}</strong>
              </p>
              {match?.message && (
                <p className="management-hint">{match.message}</p>
              )}
              <button
                className="primary"
                disabled={busy}
                onClick={() =>
                  void action(async () =>
                    onDone(
                      await Catalog.Link(local.id, preview.token),
                      "Edición vinculada.",
                    ),
                  )
                }
              >
                Vincular con registro local
              </button>
              <button
                className="quiet"
                disabled={busy}
                onClick={() => onDone(local, "")}
              >
                Abrir registro existente
              </button>
            </>
          ) : (
            <>
              <div className="add-status">
                <button
                  className="primary"
                  disabled={busy}
                  onClick={() => add(status)}
                >
                  Agregar ·{" "}
                  {info.statuses.find((item) => item.value === status)?.label}
                </button>
                <select
                  aria-label="Estado al agregar"
                  disabled={busy}
                  value={status}
                  onChange={(event) => add(event.target.value)}
                >
                  {info.statuses.map((item) => (
                    <option key={item.value} value={item.value}>
                      {item.label}
                    </option>
                  ))}
                </select>
              </div>
              {!!snapshot.work.series?.length && (
                <label className="catalog-series">
                  Serie
                  <select
                    aria-label="Serie al agregar"
                    value={series}
                    disabled={busy}
                    onChange={(event) => setSeries(event.target.value)}
                  >
                    <option value="">Sin serie</option>
                    {snapshot.work.series.map((item) => (
                      <option key={item.id} value={item.id}>
                        {item.name}
                        {item.position != null ? ` · #${item.position}` : ""}
                      </option>
                    ))}
                  </select>
                </label>
              )}
            </>
          )}
          <button
            className="secondary"
            disabled={busy}
            onClick={() => setPicking((value) => !value)}
          >
            Elegir otra edición
          </button>
          {busy && <p role="status">Guardando o consultando…</p>}
          {error && (
            <p className="error-banner" role="alert">
              {error}
            </p>
          )}
        </>
      }
    >
      {picking && (
        <EditionPicker
          configured={configured}
          workID={snapshot.work.id}
          disabled={busy}
          onClose={() => setPicking(false)}
          onWork={openWork}
          onEdition={(id) =>
            void action(async () => accept(await Catalog.Preview(id)))
          }
        />
      )}
      {!!snapshot.warnings?.length && (
        <ul className="catalog-warnings">
          {snapshot.warnings.map((warning, index) => (
            <li key={index}>{warning}</li>
          ))}
        </ul>
      )}
    </BookProfile>
  );
}
