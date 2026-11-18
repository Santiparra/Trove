import { useEffect, useRef, useState } from "react";
import * as Catalog from "../../bindings/github.com/Santiparra/Trove/internal/desktop/catalog";
import type {
  HealthReport,
  ProviderStatus,
} from "../../bindings/github.com/Santiparra/Trove/internal/providers/models";
import { errorMessage } from "../books/presentation";

const labels: Record<string, string> = {
  unchecked: "Sin comprobar",
  unconfigured: "Sin configurar",
  healthy: "Disponible",
  degraded: "Con problemas",
  paused: "En pausa",
  "access-denied": "Acceso rechazado",
};
function summary(report: HealthReport | null) {
  if (report?.localOnly) return { label: "Solo local", tone: "unknown" };
  if (!report?.providers.length)
    return { label: "Sin comprobar", tone: "unknown" };
  if (report.checking) return { label: "Comprobando…", tone: "unknown" };
  const problems = report.providers.filter((item) =>
    ["degraded", "paused", "access-denied"].includes(item.state),
  );
  if (problems.length) {
    if (problems.every((item) => item.state === "access-denied"))
      return { label: "Revisar acceso", tone: "warning" };
    return {
      label:
        problems.length === report.providers.length
          ? "Sin servicio"
          : `${problems.length} con problemas`,
      tone: "warning",
    };
  }
  if (report.providers.some((item) => item.state === "unchecked"))
    return { label: "Sin comprobar", tone: "unknown" };
  const ready = report.providers.filter(
    (item) => item.state === "healthy",
  ).length;
  if (ready !== report.providers.length)
    return {
      label: `${ready} disponible${ready === 1 ? "" : "s"}`,
      tone: "unknown",
    };
  return { label: "OK", tone: "healthy" };
}
function statusText(provider: ProviderStatus) {
  return provider.checking
    ? "Consultando…"
    : labels[provider.state] || "Sin comprobar";
}

export default function ProviderHealth({
  disabled,
  configured,
  onLocalOnly,
}: {
  disabled: boolean;
  configured: boolean;
  onLocalOnly?: (value: boolean) => void;
}) {
  const [report, setReport] = useState<HealthReport | null>(null);
  const [open, setOpen] = useState(false);
  const [checking, setChecking] = useState(false);
  const [modeBusy, setModeBusy] = useState(false);
  useEffect(() => {
    if (report) onLocalOnly?.(!!report.localOnly);
  }, [report?.localOnly, onLocalOnly]);
  const [error, setError] = useState("");
  const sequence = useRef(0);
  const modeRequest = useRef(false);
  const root = useRef<HTMLDivElement>(null);
  useEffect(() => {
    let active = true;
    let reading = false;
    async function read() {
      if (reading || modeRequest.current) return;
      reading = true;
      const current = ++sequence.current;
      try {
        const value = await Catalog.ProviderHealth();
        if (active && current === sequence.current) {
          setReport(value);
          setError("");
        }
      } catch (reason) {
        if (active && current === sequence.current) {
          setReport(null);
          setError(errorMessage(reason));
        }
      } finally {
        reading = false;
      }
    }
    void read();
    // Only local IPC is refreshed; this timer never requests a provider probe.
    const timer = window.setInterval(() => {
      if (document.visibilityState !== "hidden") void read();
    }, 3000);
    return () => {
      active = false;
      clearInterval(timer);
    };
  }, [configured]);
  useEffect(() => {
    if (!open) return;
    function outside(event: PointerEvent) {
      if (!root.current?.contains(event.target as Node)) setOpen(false);
    }
    function escape(event: KeyboardEvent) {
      if (event.key === "Escape") setOpen(false);
    }
    document.addEventListener("pointerdown", outside);
    document.addEventListener("keydown", escape);
    return () => {
      document.removeEventListener("pointerdown", outside);
      document.removeEventListener("keydown", escape);
    };
  }, [open]);
  async function check() {
    if (checking) return;
    setChecking(true);
    setError("");
    const current = ++sequence.current;
    try {
      const value = await Catalog.CheckProviders();
      if (current === sequence.current) setReport(value);
    } catch (reason) {
      if (current === sequence.current) setError(errorMessage(reason));
    } finally {
      setChecking(false);
    }
  }
  async function toggleMode() {
    if (modeBusy || !report) return;
    modeRequest.current = true;
    setModeBusy(true);
    setError("");
    const current = ++sequence.current;
    try {
      const value = await Catalog.SetLocalOnly(!report.localOnly);
      if (current === sequence.current) setReport(value);
    } catch (reason) {
      if (current === sequence.current) setError(errorMessage(reason));
    } finally {
      modeRequest.current = false;
      setModeBusy(false);
    }
  }
  const state = summary(report);
  const waiting =
    !!report?.nextCheckAt && Date.parse(report.nextCheckAt) > Date.now();
  return (
    <div className="catalog-control" ref={root}>
      <button
        className={`profile-tag provider-health-${state.tone}`}
        disabled={disabled}
        aria-expanded={open}
        onClick={() => setOpen((value) => !value)}
      >
        Proveedores · {state.label}
      </button>
      {open && (
        <section
          className="catalog-key-panel provider-health-panel"
          aria-label="Salud de proveedores"
        >
          <div className="card-heading">
            <h2>Salud de proveedores</h2>
            <button
              aria-label="Cerrar salud de proveedores"
              onClick={() => setOpen(false)}
            >
              ×
            </button>
          </div>
          <button
            className="secondary"
            aria-pressed={!!report?.localOnly}
            disabled={disabled || modeBusy || !report}
            onClick={() => void toggleMode()}
          >
            Modo solo local{report?.localOnly ? " · Activado" : ""}
          </button>
          {report?.localOnly && (
            <p className="management-hint">
              Sin consultas ni descargas remotas. Se mantienen la biblioteca,
              los snapshots y las portadas cacheadas. Se conserva al reiniciar
              este perfil.
            </p>
          )}
          {report?.providers.map((provider) => (
            <article key={provider.id} className="provider-health-item">
              <strong>{provider.name}</strong>
              <span
                className={`provider-state provider-state-${provider.state}`}
              >
                {report.localOnly ? "Solo local" : statusText(provider)}
              </span>
              <p>{provider.detail || "Todavía no se consultó la API."}</p>
              {provider.checkedAt && (
                <small>
                  Última consulta:{" "}
                  {new Date(provider.checkedAt).toLocaleTimeString("es")} ·{" "}
                  {provider.latencyMS < 1 ? "<1" : provider.latencyMS} ms
                </small>
              )}
              {provider.retryAt && (
                <small>
                  Próxima prueba permitida:{" "}
                  {new Date(provider.retryAt).toLocaleTimeString("es")}
                </small>
              )}
              {provider.state === "access-denied" && (
                <small>
                  Actualizá la clave desde Claves para habilitar nuevas
                  consultas.
                </small>
              )}
            </article>
          ))}
          <button
            className="secondary"
            disabled={
              disabled ||
              checking ||
              !!report?.checking ||
              waiting ||
              !!report?.localOnly ||
              modeBusy
            }
            onClick={() => void check()}
          >
            {report?.checking || checking ? "Comprobando…" : "Comprobar ahora"}
          </button>
          {waiting && (
            <small>
              Una comprobación manual como máximo cada 15 segundos. Las pausas
              de cada proveedor se respetan.
            </small>
          )}
          <p className="management-hint">
            Estado de las APIs de catálogo, no de las portadas. Se actualiza con
            tus consultas, sin sondeos continuos a los proveedores.
          </p>
          {error && <p role="alert">{error}</p>}
        </section>
      )}
    </div>
  );
}
