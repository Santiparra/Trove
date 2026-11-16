import { useEffect, useState } from "react";
import * as Backups from "../../bindings/github.com/Santiparra/Trove/internal/desktop/backups";
import type { Status } from "../../bindings/github.com/Santiparra/Trove/internal/backup/models";
import Icon from "../components/Icon";
import BackupPanel, { phaseLabel } from "./BackupPanel";

export default function BackupControl({
  disabled,
  canOpen,
  onRestored,
}: {
  disabled: boolean;
  canOpen: () => boolean;
  onRestored: () => void;
}) {
  const [state, setState] = useState<Status | null>(null);
  const [open, setOpen] = useState(false);
  const [unavailable, setUnavailable] = useState(false);
  useEffect(() => {
    let active = true;
    const poll = () =>
      Backups.State()
        .then((value) => {
          if (active) {
            setState(value);
            setUnavailable(false);
          }
        })
        .catch(() => {
          if (active) setUnavailable(true);
        });
    void poll();
    const timer = setInterval(() => void poll(), 1500);
    return () => {
      active = false;
      clearInterval(timer);
    };
  }, []);
  return (
    <>
      <button
        className={`backup-indicator backup-${unavailable ? "error" : state?.phase || "unconfigured"}`}
        disabled={disabled}
        onClick={() => {
          if (canOpen()) setOpen(true);
        }}
        aria-label="Respaldo y restauración"
        title={
          unavailable
            ? "No se pudo consultar el estado del respaldo"
            : state?.lastError || phaseLabel(state?.phase || "unconfigured")
        }
      >
        <Icon name="backup" size={15} />
        <span>
          {unavailable
            ? "Respaldo no disponible"
            : phaseLabel(state?.phase || "unconfigured")}
        </span>
      </button>
      {open && (
        <BackupPanel
          state={state}
          onState={setState}
          onClose={() => setOpen(false)}
          onRestored={onRestored}
        />
      )}
    </>
  );
}
