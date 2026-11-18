import { useEffect, useRef, useState } from "react";
import Icon from "./Icon";

export default function SuccessToast({
  message,
  onDismiss,
}: {
  message: string;
  onDismiss: () => void;
}) {
  const [paused, setPaused] = useState(false);
  const dismiss = useRef(onDismiss);
  dismiss.current = onDismiss;
  useEffect(() => {
    if (paused) return;
    const timer = window.setTimeout(() => dismiss.current(), 5000);
    return () => clearTimeout(timer);
  }, [paused]);
  return (
    <div
      className="success-toast"
      role="status"
      aria-live="polite"
      aria-atomic="true"
      onPointerEnter={() => setPaused(true)}
      onPointerLeave={() => setPaused(false)}
      onFocus={() => setPaused(true)}
      onBlur={(event) => {
        if (!event.currentTarget.contains(event.relatedTarget))
          setPaused(false);
      }}
    >
      <Icon name="check" size={18} />
      <span>{message}</span>
      <button aria-label="Cerrar aviso" onClick={onDismiss}>
        ×
      </button>
    </div>
  );
}
