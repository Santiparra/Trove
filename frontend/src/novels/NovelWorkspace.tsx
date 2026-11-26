import { useEffect, useRef, useState } from "react";
import * as Novels from "../../bindings/github.com/Santiparra/Trove/internal/desktop/novels";
import type {
  Novel,
  Volume,
  Input,
  Dashboard,
} from "../../bindings/github.com/Santiparra/Trove/internal/novels/models";
import type { Book } from "../../bindings/github.com/Santiparra/Trove/internal/books/models";
import type {
  Info,
  NovelPreview,
} from "../../bindings/github.com/Santiparra/Trove/internal/desktop/models";
import BookProfile from "../books/BookProfile";
import ReadingHistory from "../books/ReadingHistory";
import { dateLabel, errorMessage } from "../books/presentation";
import NovelEditor from "./NovelEditor";
import VolumeEditor from "./VolumeEditor";
import Icon from "../components/Icon";
import type {
  Candidate,
  SearchResult,
} from "../../bindings/github.com/Santiparra/Trove/internal/novelcatalog/models";
import NovelSearch from "./NovelSearch";
import CatalogNovel, { PublicationSources } from "./CatalogNovel";
import {
  representedNovel,
  volumeLabel,
  novelColors,
  novelCoverURL,
} from "./presentation";
import NovelPortal, { NovelCover } from "./NovelPortal";
import VolumeGallery from "./VolumeGallery";
import DeleteDialog from "../books/DeleteDialog";
import {
  Series,
  Publication,
} from "../../bindings/github.com/Santiparra/Trove/internal/novelcatalog/models";

function progress(n: Novel) {
  return `${n.completed ?? "?"}/${progressTotal(n) ?? "?"}`;
}
function currentRun(n: Novel) {
  return n.readings?.find((r) => r.id === n.activeRunID);
}
function progressTotal(n: Novel) {
  return currentRun(n)?.volumeIDs.length ?? n.total;
}
function lastNumber(n: Novel) {
  const run = currentRun(n);
  if (run)
    return (
      n.volumes.find(
        (v) => v.id === run.volumeIDs[Math.max(0, (n.completed || 0) - 1)],
      )?.number || 0
    );
  return n.completed || 0;
}
function nextNumber(n: Novel) {
  if (n.activeVolumeID)
    return n.volumes.find((v) => v.id === n.activeVolumeID)?.number || 0;
  const run = currentRun(n);
  if (run)
    return n.status === "read" || n.completed == null
      ? 0
      : n.volumes.find((v) => v.id === run.volumeIDs[n.completed!])?.number ||
          0;
  return n.completed == null ||
    n.status === "read" ||
    (n.total != null && n.completed >= n.total)
    ? 0
    : n.completed + 1;
}
function asInput(n: Novel): Input {
  return {
    id: n.id,
    title: n.title,
    authors: n.authors,
    status: n.status,
    total: n.total,
    completed: n.completed,
    notes: n.notes,
    favorite: n.favorite,
    addedOn: n.addedOn,
  };
}

export default function NovelWorkspace({
  info,
  onBusy,
  onDirty,
  canLeave,
  onNotice,
  localOnly = false,
}: {
  info: Info;
  onBusy: (value: boolean) => void;
  onDirty: (value: boolean) => void;
  canLeave: () => boolean;
  onNotice: (message: string) => void;
  localOnly?: boolean;
}) {
  const [items, setItems] = useState<Novel[]>([]);
  const [selected, setSelected] = useState<Novel | null>(null);
  const [query, setQuery] = useState("");
  const [status, setStatus] = useState("");
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [revision, setRevision] = useState(0);
  const [editing, setEditing] = useState(false);
  const [volumeEdit, setVolumeEdit] = useState<{
    volume: Volume | null;
    number: number;
  } | null>(null);
  const [notes, setNotes] = useState("");
  const [readingDirty, setReadingDirty] = useState(false);
  const [runDirty, setRunDirty] = useState(false);
  const [historyRevision, setHistoryRevision] = useState(0);
  const [rereadSelection, setRereadSelection] = useState<string[] | null>(null);
  const [all, setAll] = useState(false);
  const [dashboard, setDashboard] = useState<Dashboard | null>(null);
  const [portalError, setPortalError] = useState("");
  const [confirmation, setConfirmation] = useState<
    { kind: "delete" } | { kind: "complete"; count: number } | null
  >(null);
  useEffect(() => {
    let active = true;
    setPortalError("");
    Novels.Dashboard()
      .then((value) => {
        if (active) setDashboard(value);
      })
      .catch((reason) => {
        if (active) setPortalError(errorMessage(reason));
      });
    return () => {
      active = false;
    };
  }, [revision]);
  const [searchText, setSearchText] = useState("");
  const [authorID, setAuthorID] = useState("");
  const [remote, setRemote] = useState<SearchResult | null>(null);
  const [remoteLoading, setRemoteLoading] = useState(false);
  const [remoteError, setRemoteError] = useState("");
  const [page, setPage] = useState(1);
  const [preview, setPreview] = useState<NovelPreview | null>(null);
  const [browseLoading, setBrowseLoading] = useState(false);
  const [browseError, setBrowseError] = useState("");
  const browseRequest = useRef(0);
  useEffect(
    () => () => {
      browseRequest.current++;
    },
    [],
  );
  const [linkTarget, setLinkTarget] = useState<Novel | null>(null);
  const [linking, setLinking] = useState(false);
  const [linkQuery, setLinkQuery] = useState("");
  const [viewedNumber, setViewedNumber] = useState<number | null>(null);
  const [metadataError, setMetadataError] = useState("");
  const [metadataLoading, setMetadataLoading] = useState(false);
  const shownNumber = selected
    ? (viewedNumber ??
      (nextNumber(selected) ||
        lastNumber(selected) ||
        selected.volumes[selected.volumes.length - 1]?.number ||
        0))
    : 0;
  useEffect(() => {
    let active = true;
    setRemote(null);
    setRemoteError("");
    setRemoteLoading(false);
    if (
      selected ||
      preview ||
      editing ||
      volumeEdit ||
      localOnly ||
      query.trim().length < 2 ||
      status
    )
      return;
    setRemoteLoading(true);
    const timer = setTimeout(() => {
      Novels.SearchCatalog(query, authorID, page, "list")
        .then((value) => {
          if (active) setRemote(value);
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
  }, [
    query,
    authorID,
    page,
    status,
    selected?.id,
    !!preview,
    editing,
    volumeEdit,
    localOnly,
    revision,
  ]);
  useEffect(() => {
    let active = true;
    setMetadataError("");
    setMetadataLoading(false);
    if (
      !selected?.catalog ||
      !shownNumber ||
      selected.volumes.some((v) => v.number === shownNumber && v.catalog) ||
      preview ||
      editing ||
      volumeEdit ||
      localOnly
    )
      return;
    setMetadataLoading(true);
    Novels.RefreshCatalog(selected.id, shownNumber, false)
      .then((value) => {
        if (active) setSelected(value);
      })
      .catch((reason) => {
        if (active) setMetadataError(errorMessage(reason));
      })
      .finally(() => {
        if (active) setMetadataLoading(false);
      });
    return () => {
      active = false;
    };
  }, [
    selected?.id,
    selected?.catalog?.id,
    shownNumber,
    !!preview,
    editing,
    volumeEdit,
    localOnly,
  ]);
  useEffect(() => {
    let active = true;
    setLoading(true);
    setError("");
    const timer = setTimeout(
      () => {
        Novels.List(query, status)
          .then((value) => {
            if (active) setItems(value);
          })
          .catch((reason) => {
            if (active) setError(errorMessage(reason));
          })
          .finally(() => {
            if (active) setLoading(false);
          });
      },
      query ? 140 : 0,
    );
    return () => {
      active = false;
      clearTimeout(timer);
    };
  }, [query, status, revision]);
  useEffect(
    () => setNotes(selected?.notes || ""),
    [selected?.id, selected?.notes],
  );
  useEffect(() => {
    if (!editing && !volumeEdit)
      onDirty(
        (!!selected && notes !== selected.notes) || readingDirty || runDirty,
      );
  }, [
    notes,
    selected?.notes,
    selected?.id,
    editing,
    volumeEdit,
    readingDirty,
    runDirty,
    onDirty,
  ]);
  function leave(action: () => void) {
    if (busy || !canLeave()) return;
    if (readingDirty || runDirty) {
      setHistoryRevision((v) => v + 1);
      setReadingDirty(false);
      setRunDirty(false);
    }
    onDirty(false);
    setNotes(selected?.notes || "");
    setError("");
    action();
  }
  async function action(operation: () => PromiseLike<void>) {
    if (busy) return false;
    setBusy(true);
    onBusy(true);
    setError("");
    try {
      await operation();
      return true;
    } catch (reason) {
      setError(errorMessage(reason));
      return false;
    } finally {
      setBusy(false);
      onBusy(false);
    }
  }
  function commit(n: Novel, message: string, showCurrent = false) {
    setSelected(n);
    setRereadSelection(null);
    if (showCurrent) setViewedNumber(null);
    setEditing(false);
    setVolumeEdit(null);
    setRevision((value) => value + 1);
    onDirty(!editing && !volumeEdit && notes !== n.notes);
    onNotice(message);
  }
  function open(n: Novel) {
    leave(
      () =>
        void action(async () => {
          setSelected(await Novels.Get(n.id));
          setViewedNumber(null);
          setPreview(null);
          setLinking(false);
        }),
    );
  }
  function openByID(id: string, number?: number) {
    leave(
      () =>
        void action(async () => {
          setSelected(await Novels.Get(id));
          setViewedNumber(number ?? null);
          setPreview(null);
          setLinking(false);
        }),
    );
  }
  function list(statusValue: string) {
    leave(() => {
      setAll(true);
      setStatus(statusValue);
      setQuery("");
      setSearchText("");
      setAuthorID("");
      setPage(1);
    });
  }
  function adjustProgress(value: number) {
    if (!selected || busy) return;
    if (
      selected.total != null &&
      value === selected.total &&
      selected.status !== "read"
    ) {
      setConfirmation({ kind: "complete", count: value });
      return;
    }
    void action(async () =>
      commit(
        await Novels.SetCompleted(selected.id, value, false),
        "Progreso actualizado sin modificar fechas.",
        true,
      ),
    );
  }
  function search(value: string) {
    leave(() => {
      setAuthorID(
        selected?.catalog?.authors.find((author) => author.name === value)
          ?.id || "",
      );
      setSelected(null);
      setPreview(null);
      browseRequest.current++;
      setBrowseLoading(false);
      setLinking(false);
      setViewedNumber(null);
      setQuery(value);
      setSearchText(value);
      setPage(1);
      setStatus("");
      setAll(true);
    });
  }
  function allResults(value: string) {
    leave(() => {
      setQuery(value);
      setSearchText(value);
      setPage(1);
      setAll(true);
      setStatus("");
    });
  }
  function candidate(item: Candidate, target: Novel | null = null) {
    leave(() => {
      const request = ++browseRequest.current;
      setBrowseLoading(true);
      setBrowseError("");
      setLinkTarget(target);
      setLinking(false);
      const preferred = target
        ? nextNumber(target) || target.completed || 1
        : 1;
      setPreview({
        token: "",
        series: new Series({
          id: item.id,
          title: item.title,
          originalTitle: item.originalTitle,
          coverURL: item.coverURL,
        }),
        publication: new Publication({ number: preferred }),
        match: { kind: "new", novel: null, message: "" },
      });
      void Novels.BrowseCatalog(
        item.id,
        item.matchedBookID || "",
        item.matchedReleaseID || "",
      )
        .then((value) => {
          if (request !== browseRequest.current) return;
          if (
            value.match.kind === "existing" &&
            value.match.novel &&
            !value.token
          ) {
            setSelected(value.match.novel);
            setViewedNumber(
              item.matchedBookID
                ? value.series.books.find(
                    (ref) => ref.id === item.matchedBookID,
                  )?.number || null
                : null,
            );
            setPreview(null);
            setLinking(false);
            return;
          }
          setLinkTarget(
            target ||
              (value.match.kind === "existing" ? value.match.novel : null),
          );
          const number = item.matchedBookID
            ? value.series.books.find((ref) => ref.id === item.matchedBookID)
                ?.number || 1
            : Math.min(preferred, value.series.books.length) || 1;
          setPreview({
            ...value,
            publication: value.publication.bookID
              ? value.publication
              : new Publication({
                  number,
                  releaseID: item.matchedReleaseID || "",
                }),
          });
          setLinking(false);
        })
        .catch((reason) => {
          if (request === browseRequest.current)
            setBrowseError(errorMessage(reason));
        })
        .finally(() => {
          if (request === browseRequest.current) setBrowseLoading(false);
        });
    });
  }
  function choosePublication() {
    if (!selected) return;
    leave(
      () =>
        void action(async () => {
          setLinkTarget(selected);
          setPreview(
            await Novels.PrepareCatalog(selected.id, shownNumber || 1),
          );
        }),
    );
  }
  if (preview)
    return (
      <CatalogNovel
        key={preview.token}
        initial={preview}
        target={linkTarget}
        info={info}
        localOnly={localOnly}
        loadingSeries={browseLoading}
        browseError={browseError}
        onBusy={onBusy}
        onSearch={search}
        onBack={() => {
          browseRequest.current++;
          setBrowseLoading(false);
          setBrowseError("");
          setPreview(null);
        }}
        onDone={(n, message, number) => {
          setSelected(n);
          setPreview(null);
          setLinkTarget(null);
          setLinking(false);
          setViewedNumber(number || null);
          onDirty(false);
          setRevision((value) => value + 1);
          if (message) onNotice(message);
        }}
      />
    );
  if (editing)
    return (
      <NovelEditor
        key={selected?.id || "new"}
        novel={selected}
        info={info}
        busy={busy}
        error={error}
        onDirty={onDirty}
        onCancel={() => leave(() => setEditing(false))}
        onSave={(input) =>
          void action(async () =>
            commit(
              await Novels.Save(input),
              "Novela guardada en la biblioteca local.",
            ),
          )
        }
      />
    );
  if (volumeEdit && selected)
    return (
      <VolumeEditor
        key={volumeEdit.volume?.id || `new-${volumeEdit.number}`}
        volume={volumeEdit.volume}
        number={volumeEdit.number}
        total={selected.total}
        info={info}
        busy={busy}
        error={error}
        onDirty={onDirty}
        onCancel={() => leave(() => setVolumeEdit(null))}
        onSave={(input) =>
          void action(async () =>
            commit(
              await Novels.SaveVolume(selected.id, input),
              "Datos del volumen guardados.",
            ),
          )
        }
      />
    );
  if (selected) {
    const run = currentRun(selected);
    const rereadable = selected.volumes.filter(
      (v) =>
        v.cycle.status === "read" ||
        v.history?.some((c) => c.status === "read"),
    );
    const next = nextNumber(selected);
    const shown = shownNumber;
    const current = selected.volumes.find((volume) => volume.number === shown);
    const shownLabel = volumeLabel(selected, shown);
    const nextLabel = next ? volumeLabel(selected, next) : "";
    const registeredNumbers = new Set(
      selected.volumes.map((volume) => volume.number),
    );
    let numberToRegister = next || 1;
    while (registeredNumbers.has(numberToRegister)) numberToRegister++;
    if (selected.total != null && numberToRegister > selected.total) {
      numberToRegister = 1;
      while (registeredNumbers.has(numberToRegister)) numberToRegister++;
    }
    const book: Book = {
      id: selected.id,
      medium: "light-novels",
      title: selected.title,
      authors: selected.authors.length
        ? selected.authors
        : selected.catalog?.authors.map((author) => author.name) || [],
      series: null,
      edition: current?.edition || {
        id: "",
        language: "",
        format: "",
        identifier: "",
        owned: false,
        identifiers: [],
        catalog: null,
      },
      cycle: current?.cycle || {
        id: "",
        status: selected.status,
        startedOn: "",
        finishedOn: "",
      },
      notes: selected.notes,
      addedOn: selected.addedOn,
      favorite: selected.favorite,
    };
    return (
      <>
        <BookProfile
          sourcesBeforeChildren
          compactManagement
          loading={metadataLoading}
          metadataActions={
            <section
              className="metadata-actions"
              aria-label="Gestión de metadata"
            >
              <h3>Catálogo y metadata</h3>
              <div className="catalog-actions">
                {selected.catalog ? (
                  <>
                    <button
                      className="secondary"
                      disabled={busy || localOnly || metadataLoading}
                      onClick={() =>
                        void action(async () => {
                          const updated = await Novels.RefreshCatalog(
                            selected.id,
                            shown || 0,
                            true,
                          );
                          setMetadataError("");
                          commit(updated, "Metadata de novela actualizada.");
                        })
                      }
                    >
                      Refrescar catálogo
                    </button>
                    <button
                      className="quiet danger-text"
                      disabled={busy}
                      onClick={() => {
                        if (
                          window.confirm(
                            "¿Quitar el vínculo y sus snapshots? Se conservan progreso, fechas y datos personales.",
                          )
                        )
                          void action(async () =>
                            commit(
                              await Novels.UnlinkCatalog(selected.id),
                              "Catálogo desvinculado.",
                            ),
                          );
                      }}
                    >
                      Desvincular catálogo
                    </button>
                  </>
                ) : (
                  <button
                    className="secondary"
                    disabled={busy || localOnly}
                    onClick={() =>
                      leave(() => {
                        setLinkQuery(selected.title);
                        setLinking((value) => !value);
                      })
                    }
                  >
                    Vincular catálogo
                  </button>
                )}
              </div>
            </section>
          }
          book={book}
          ariaLabel="Detalle de novela"
          subtitle={`${progress(selected)} volúmenes completados${shown ? ` · Volumen ${shownLabel}${current?.title || current?.catalog?.title ? `: ${current?.title || current?.catalog?.title}` : ""}` : ""}`}
          description={
            current?.description ||
            current?.catalog?.description ||
            selected.catalog?.description
          }
          coverURL={current?.catalog?.coverURL}
          editionInfo={current?.catalog || undefined}
          formatLabel={
            current?.catalog?.releaseFormat === "print" &&
            current?.edition.format === "other"
              ? "Impreso"
              : undefined
          }
          sourceDetails={
            current?.catalog ? (
              <PublicationSources publication={current.catalog} />
            ) : undefined
          }
          busy={busy}
          onBack={() =>
            leave(() => {
              setSelected(null);
              setViewedNumber(null);
              setLinking(false);
            })
          }
          onSearch={search}
          management={
            <>
              <section
                className="management-group novel-tracking-actions"
                aria-label="Seguimiento de novela"
              >
                <h3>Seguimiento</h3>
                {run && (
                  <span className="management-hint">
                    Recorrido de relectura · {run.volumeIDs.length} volúmenes
                  </span>
                )}
                {selected.activeVolumeID && (
                  <span className="management-hint">
                    Relectura individual en curso
                  </span>
                )}
                <select
                  className="novel-status-select"
                  style={{
                    color: novelColors[selected.status],
                    borderColor: novelColors[selected.status],
                  }}
                  aria-label={`Estado de ${selected.title}`}
                  value={selected.status}
                  disabled={busy}
                  onChange={(event) =>
                    void action(async () =>
                      commit(
                        await Novels.ChangeStatus(
                          selected.id,
                          event.target.value,
                        ),
                        "Estado de la novela actualizado.",
                      ),
                    )
                  }
                >
                  {info.statuses.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.value === "read" ? "Leída" : option.label}
                    </option>
                  ))}
                </select>
                <div
                  className="novel-progress-control"
                  role="group"
                  aria-label={`Progreso: ${progress(selected)}`}
                >
                  <strong className="novel-progress">
                    {selected.completed ?? "?"}
                  </strong>
                  <span className="novel-progress-arrows">
                    <button
                      aria-label="Aumentar volúmenes completados"
                      title="Aumentar progreso sin cambiar fechas"
                      disabled={
                        busy ||
                        !!run ||
                        !!selected.activeVolumeID ||
                        selected.completed == null ||
                        (selected.total != null &&
                          selected.completed >= selected.total)
                      }
                      onClick={() => adjustProgress(selected.completed! + 1)}
                    >
                      ▴
                    </button>
                    <button
                      aria-label="Reducir volúmenes completados"
                      title="Reducir progreso sin cambiar fechas"
                      disabled={
                        busy ||
                        !!run ||
                        !!selected.activeVolumeID ||
                        selected.completed == null ||
                        selected.completed <= 0
                      }
                      onClick={() => adjustProgress(selected.completed! - 1)}
                    >
                      ▾
                    </button>
                  </span>
                  <strong className="novel-progress">
                    /{progressTotal(selected) ?? "?"}
                  </strong>
                </div>
                <span className="management-hint">Volúmenes completados</span>
                {!!rereadable.length && (
                  <button
                    className="secondary"
                    disabled={
                      busy ||
                      !!selected.activeVolumeID ||
                      selected.status === "reading" ||
                      (!!run &&
                        !["read", "did-not-finish"].includes(selected.status))
                    }
                    onClick={() =>
                      leave(() =>
                        setRereadSelection(rereadable.map((v) => v.id)),
                      )
                    }
                  >
                    Releer novela
                  </button>
                )}
                {viewedNumber != null && !!next && viewedNumber !== next && (
                  <button
                    className="quiet"
                    disabled={busy}
                    onClick={() => setViewedNumber(null)}
                  >
                    Volver al volumen actual {nextLabel}
                  </button>
                )}
                {selected.total != null &&
                  selected.completed === selected.total &&
                  selected.status !== "read" && (
                    <button
                      className="primary"
                      disabled={busy}
                      onClick={() =>
                        void action(async () =>
                          commit(
                            await Novels.ChangeStatus(selected.id, "read"),
                            "Novela marcada como leída.",
                          ),
                        )
                      }
                    >
                      Marcar novela como leída
                    </button>
                  )}
                <button
                  className="primary"
                  disabled={
                    busy ||
                    !next ||
                    (selected.status === "reading" &&
                      current?.cycle.status === "reading")
                  }
                  onClick={() =>
                    void action(async () =>
                      commit(
                        await Novels.StartVolume(selected.id),
                        `Volumen ${next} iniciado.`,
                        true,
                      ),
                    )
                  }
                >
                  Empezar volumen{next ? ` ${nextLabel}` : ""}
                </button>
                <button
                  className="secondary"
                  disabled={busy || !next}
                  onClick={() =>
                    void action(async () =>
                      commit(
                        await Novels.FinishVolume(selected.id),
                        "Volumen terminado. Progreso actualizado.",
                        true,
                      ),
                    )
                  }
                >
                  {next && selected.total === next
                    ? "Terminar último volumen y marcar leída"
                    : "Volumen terminado"}
                </button>
              </section>
              <section
                className="management-group"
                aria-label="Gestión de novela"
              >
                <h3>Novela</h3>
                <div className="novel-record-buttons">
                  <button
                    className="quiet"
                    disabled={busy}
                    onClick={() => leave(() => setEditing(true))}
                    aria-label="Editar novela y progreso"
                  >
                    Editar novela
                  </button>
                  <button
                    className={`flag-button ${selected.favorite ? "is-active" : ""}`}
                    aria-pressed={selected.favorite}
                    disabled={busy}
                    onClick={() =>
                      void action(async () =>
                        commit(
                          await Novels.Save({
                            ...asInput(selected),
                            favorite: !selected.favorite,
                          }),
                          "Favorita actualizada.",
                        ),
                      )
                    }
                  >
                    Favorita
                  </button>
                  <button
                    className="quiet danger-text"
                    disabled={busy}
                    aria-label="Eliminar novela"
                    onClick={() => {
                      if (canLeave()) {
                        setError("");
                        setConfirmation({ kind: "delete" });
                      }
                    }}
                  >
                    Eliminar
                  </button>
                  <small>Alta: {dateLabel(selected.addedOn)}</small>
                </div>
              </section>
              <section
                className="management-group"
                aria-label="Gestión de volumen"
              >
                <h3>Volumen {shownLabel}</h3>
                <div className="novel-volume-buttons">
                  {shown > 0 && (
                    <button
                      className="secondary"
                      disabled={busy || metadataLoading}
                      aria-label={`Editar datos del volumen ${shownLabel}`}
                      onClick={() =>
                        leave(() =>
                          setVolumeEdit({
                            volume: current || null,
                            number: shown,
                          }),
                        )
                      }
                    >
                      Editar volumen
                    </button>
                  )}
                  {selected.catalog && (
                    <button
                      className="secondary"
                      disabled={busy || localOnly}
                      onClick={choosePublication}
                    >
                      Elegir volumen o publicación
                    </button>
                  )}
                </div>
                {metadataLoading && (
                  <small role="status">Consultando metadata del volumen…</small>
                )}
                {metadataError && (
                  <p className="management-hint" role="alert">
                    {metadataError}
                  </p>
                )}
              </section>
              {error && !confirmation && (
                <p className="error-banner" role="alert">
                  {error}
                </p>
              )}
            </>
          }
        >
          {rereadSelection && (
            <section
              className="portal-card reread-selection"
              aria-label="Nuevo recorrido de novela"
            >
              <h2>Releer novela</h2>
              <p>
                Elegí los volúmenes del nuevo recorrido. Sus lecturas anteriores
                se conservan; se iniciarán cuando elijas Empezar volumen.
              </p>
              {rereadable.map((v) => (
                <label key={v.id}>
                  <input
                    type="checkbox"
                    disabled={busy}
                    checked={rereadSelection.includes(v.id)}
                    onChange={(e) =>
                      setRereadSelection((old) =>
                        e.target.checked
                          ? [...(old || []), v.id]
                          : (old || []).filter((id) => id !== v.id),
                      )
                    }
                  />
                  Volumen {volumeLabel(selected, v.number)}
                  {v.title ? ` · ${v.title}` : ""}
                </label>
              ))}
              <div className="catalog-actions">
                <button
                  className="primary"
                  disabled={busy || !rereadSelection.length}
                  onClick={() =>
                    void action(async () =>
                      commit(
                        await Novels.RereadNovel(selected.id, rereadSelection),
                        "Nuevo recorrido preparado. El historial anterior se conserva.",
                        true,
                      ),
                    )
                  }
                >
                  Crear recorrido
                </button>
                <button
                  className="quiet"
                  disabled={busy}
                  onClick={() => setRereadSelection(null)}
                >
                  Cancelar
                </button>
              </div>
            </section>
          )}
          {run && (
            <ReadingHistory
              key={`${selected.id}-runs-${historyRevision}`}
              title="Recorridos de novela"
              current={run.cycle}
              history={selected.readings
                ?.filter((r) => r.id !== run.id)
                .map((r) => r.cycle)}
              labels={Object.fromEntries(
                (selected.readings || []).map((r, i) => [
                  r.id,
                  `${i === 0 ? "Lectura inicial" : `Recorrido ${i + 1}`} · ${r.completed}/${r.volumeIDs.length} volúmenes · ${r.volumeIDs
                    .map((id) => {
                      const v = selected.volumes.find((v) => v.id === id);
                      return v ? volumeLabel(selected, v.number) : "";
                    })
                    .join(", ")}`,
                ]),
              )}
              info={info}
              busy={busy}
              allowAdd={false}
              onDirty={setRunDirty}
              onSave={(input) =>
                action(async () =>
                  commit(
                    await Novels.SaveRun(selected.id, input),
                    "Recorrido corregido.",
                  ),
                )
              }
            />
          )}
          {current && (
            <ReadingHistory
              key={`${selected.id}-${current.id}-${historyRevision}`}
              title={`Lecturas del volumen ${shownLabel}`}
              current={current.cycle}
              history={current.history}
              info={info}
              busy={busy}
              onDirty={setReadingDirty}
              onReread={
                !selected.activeVolumeID &&
                selected.status !== "reading" &&
                (!run || ["read", "did-not-finish"].includes(selected.status))
                  ? () =>
                      void action(async () =>
                        commit(
                          await Novels.RereadVolume(selected.id, current.id),
                          "Relectura de volumen iniciada.",
                          true,
                        ),
                      )
                  : undefined
              }
              onSave={(input) =>
                action(async () =>
                  commit(
                    await Novels.SaveReading(selected.id, current.id, input),
                    "Lectura del volumen guardada.",
                  ),
                )
              }
            />
          )}
          {linking && (
            <section className="edition-picker" aria-label="Vincular novela">
              <div className="card-heading">
                <h2>Buscar novela para vincular</h2>
                <button className="quiet" onClick={() => setLinking(false)}>
                  Cerrar búsqueda
                </button>
              </div>
              <NovelSearch
                value={linkQuery}
                onChange={setLinkQuery}
                disabled={busy}
                localOnly={localOnly}
                catalogOnly
                onAll={() => {}}
                onNovel={() => {}}
                onCandidate={(item) => candidate(item, selected)}
              />
            </section>
          )}
          {!!current?.catalog?.warnings.length && (
            <ul className="catalog-warnings">
              {current.catalog.warnings.map((warning, index) => (
                <li key={index}>{warning}</li>
              ))}
            </ul>
          )}
          <section
            className="novel-volumes"
            aria-label="Volúmenes de la novela"
          >
            <div className="card-heading">
              <h2>Volúmenes</h2>
              <button
                className="secondary"
                disabled={
                  busy ||
                  (selected.total != null &&
                    selected.volumes.length >= selected.total)
                }
                onClick={() =>
                  leave(() =>
                    setVolumeEdit({
                      volume: null,
                      number: numberToRegister,
                    }),
                  )
                }
              >
                Registrar volumen
              </button>
            </div>
            {!selected.volumes.length && (
              <p className="management-hint">
                Sin volúmenes registrados. Empezar, terminar o declarar un
                progreso conocido crea sus registros; una cantidad desconocida
                no inventa volúmenes.
              </p>
            )}
            <VolumeGallery
              novel={selected}
              busy={busy}
              next={
                run || selected.activeVolumeID
                  ? next
                  : selected.completed != null
                    ? selected.completed + 1
                    : next
              }
              onVolume={(number) =>
                leave(() => {
                  if (
                    selected.volumes.some((volume) => volume.number === number)
                  ) {
                    setViewedNumber(number);
                  } else if (!localOnly) {
                    void action(async () => {
                      setLinkTarget(selected);
                      setPreview(
                        await Novels.PrepareCatalog(selected.id, number),
                      );
                    });
                  } else {
                    setViewedNumber(number);
                  }
                })
              }
            />
          </section>
          <section className="portal-card profile-notes">
            <div className="card-heading">
              <h2>Notas de la novela</h2>
              <button
                className="small-save"
                disabled={busy || notes === selected.notes}
                onClick={() =>
                  void action(async () =>
                    commit(
                      await Novels.Save({ ...asInput(selected), notes }),
                      "Notas de la novela guardadas.",
                    ),
                  )
                }
              >
                Guardar notas
              </button>
            </div>
            <label>
              <span className="sr-only">Notas de la novela</span>
              <textarea
                rows={4}
                disabled={busy}
                value={notes}
                onChange={(event) => setNotes(event.target.value)}
              />
            </label>
          </section>
        </BookProfile>
        {confirmation && (
          <DeleteDialog
            book={book}
            busy={busy}
            error={error}
            heading={
              confirmation.kind === "delete"
                ? "¿Eliminar esta novela?"
                : "¿Marcar la novela como leída?"
            }
            description={
              confirmation.kind === "delete" ? (
                <>
                  <strong>{selected.title}</strong> y todos sus volúmenes se
                  eliminarán junto a sus notas y seguimiento. No afecta a otras
                  novelas ni a Libros.
                </>
              ) : (
                <>
                  El progreso llegará a {selected.total}/{selected.total}. Se
                  marcará Leída sin inventar ni cambiar fechas de lectura.
                </>
              )
            }
            danger={confirmation.kind === "delete"}
            confirmLabel={
              confirmation.kind === "delete"
                ? "Eliminar definitivamente"
                : "Completar y marcar leída"
            }
            onCancel={() => {
              if (!busy) {
                setConfirmation(null);
                setError("");
              }
            }}
            onConfirm={() =>
              void action(async () => {
                if (confirmation.kind === "delete") {
                  await Novels.Delete(selected.id);
                  setSelected(null);
                  setRevision((value) => value + 1);
                  onDirty(false);
                  onNotice("Novela y volúmenes eliminados.");
                } else {
                  commit(
                    await Novels.SetCompleted(
                      selected.id,
                      confirmation.count,
                      true,
                    ),
                    "Progreso completado sin modificar fechas.",
                    true,
                  );
                }
                setConfirmation(null);
              })
            }
          />
        )}
      </>
    );
  }
  const portal = !all && !query && !status;
  const visible = items;
  return (
    <section className="novel-workspace" aria-label="Biblioteca de novelas">
      {!portal && (
        <div className="novel-list-navigation">
          <button
            className="back-link"
            disabled={busy}
            onClick={() => {
              setAll(false);
              setStatus("");
              setQuery("");
              setSearchText("");
              setAuthorID("");
              setPage(1);
            }}
          >
            <Icon name="arrow" size={17} />
            Volver al portal
          </button>
        </div>
      )}
      <div className="page-heading">
        <h1>{all || query || status ? "Novelas ligeras" : "Tus novelas"}</h1>
        <button
          className="primary"
          disabled={busy}
          onClick={() =>
            leave(() => {
              setSelected(null);
              setEditing(true);
            })
          }
        >
          <Icon name="plus" size={17} /> Añadir novela
        </button>
      </div>
      <div className="library-toolbar">
        <NovelSearch
          value={searchText}
          onChange={(value) => {
            setSearchText(value);
            setAuthorID("");
          }}
          authorID={authorID}
          disabled={busy}
          localOnly={localOnly}
          onNovel={open}
          onCandidate={(item) => candidate(item)}
          onAll={allResults}
        />
      </div>
      {portal ? (
        <>
          {portalError && <p role="alert">{portalError}</p>}
          {dashboard ? (
            <NovelPortal
              dashboard={dashboard}
              info={info}
              busy={busy}
              onList={list}
              onNovel={openByID}
            />
          ) : (
            !portalError && <p role="status">Cargando portal…</p>
          )}
        </>
      ) : (
        <>
          <div className="status-tabs" aria-label="Estados de novelas">
            <button
              className={!status ? "active" : ""}
              disabled={busy}
              onClick={() => {
                setStatus("");
                setAll(true);
              }}
            >
              Todas
            </button>
            {info.statuses.map((option) => (
              <button
                className={status === option.value ? "active" : ""}
                disabled={busy}
                key={option.value}
                style={{ color: novelColors[option.value] }}
                onClick={() => {
                  setStatus(option.value);
                  setAll(true);
                }}
              >
                {option.value === "read" ? "Leídas" : option.label}
              </button>
            ))}
          </div>
          {error && (
            <p className="error-banner" role="alert">
              {error}
              <button onClick={() => setRevision((value) => value + 1)}>
                Reintentar
              </button>
            </p>
          )}
          {loading && <p role="status">Cargando novelas…</p>}
          {!loading && !items.length && !error && (
            <div className="empty-state">
              <h2>
                {query || status
                  ? "No hay coincidencias locales"
                  : "Tu biblioteca de novelas está vacía"}
              </h2>
              {!query && !status && (
                <p>Agregá una novela y registrá sus volúmenes y progreso.</p>
              )}
            </div>
          )}
          {!loading && !!items.length && !visible.length && (
            <p className="management-hint">
              No hay novelas en curso. Podés consultar todas o elegir un estado.
            </p>
          )}
          <div className="novel-library-grid">
            {visible.map((novel) => (
              <button
                className="novel-library-card"
                disabled={busy}
                key={novel.id}
                aria-label={`Abrir novela ${novel.title}`}
                onClick={() => open(novel)}
              >
                <span className="novel-list-cover">
                  <NovelCover
                    url={novelCoverURL(novel)}
                    label={novel.title}
                    lazy
                  />
                </span>
                <div>
                  <strong>{novel.title}</strong>
                  <span>
                    {novel.authors.join(" · ") || "Autoría sin especificar"}
                  </span>
                  <span>
                    {
                      info.statuses.find(
                        (option) => option.value === novel.status,
                      )?.label
                    }
                  </span>
                </div>
                <div className="novel-card-progress">
                  <strong>{progress(novel)}</strong>
                  <small>
                    {nextNumber(novel)
                      ? `Volumen ${volumeLabel(novel, nextNumber(novel))}`
                      : novel.status === "read"
                        ? "Lectura completada"
                        : "Progreso por indicar"}
                  </small>
                </div>
              </button>
            ))}
          </div>
          {query && !status && !localOnly && (
            <section
              className="catalog-results"
              aria-label="Resultados RanobeDB"
            >
              {remoteLoading && <p role="status">Buscando en RanobeDB…</p>}
              {remoteError && (
                <p className="error-banner" role="alert">
                  {remoteError}
                  <button onClick={() => setRevision((value) => value + 1)}>
                    Reintentar catálogo
                  </button>
                </p>
              )}
              {remote?.items
                .filter((item) => !representedNovel(item, items))
                .map((item) => (
                  <button
                    className="catalog-work"
                    key={item.id}
                    disabled={busy}
                    onClick={() => candidate(item)}
                  >
                    <span>
                      <strong>{item.title}</strong>
                      <small>{item.originalTitle}</small>
                    </span>
                    <Icon name="chevron" size={17} />
                  </button>
                ))}
              {remote && !remote.items.length && (
                <p className="management-hint">
                  Sin coincidencias en RanobeDB.
                </p>
              )}
              {!!remote && remote.pages > 1 && (
                <div className="catalog-actions">
                  <button
                    className="quiet"
                    disabled={busy || page <= 1}
                    onClick={() => setPage((value) => value - 1)}
                  >
                    Anterior
                  </button>
                  <span>
                    Página {page} de {remote.pages}
                  </span>
                  <button
                    className="quiet"
                    disabled={busy || page >= remote.pages}
                    onClick={() => setPage((value) => value + 1)}
                  >
                    Siguiente
                  </button>
                </div>
              )}
            </section>
          )}
        </>
      )}
    </section>
  );
}
