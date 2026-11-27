import { useId, useState, type ReactNode } from "react";

export type ProfileTab = { label: string; content: ReactNode };

export default function ProfileTabs({ tabs }: { tabs: ProfileTab[] }) {
  const id = useId();
  const [active, setActive] = useState(0);
  return (
    <div className="novel-profile-tabs">
      <div
        role="tablist"
        aria-label="Ficha de novela"
        onKeyDown={(event) => {
          const next =
            event.key === "ArrowRight"
              ? (active + 1) % tabs.length
              : event.key === "ArrowLeft"
                ? (active + tabs.length - 1) % tabs.length
                : event.key === "Home"
                  ? 0
                  : event.key === "End"
                    ? tabs.length - 1
                    : null;
          if (next === null) return;
          event.preventDefault();
          setActive(next);
          event.currentTarget
            .querySelectorAll<HTMLButtonElement>("[role=tab]")
            [next]?.focus();
        }}
      >
        {tabs.map(({ label }, i) => (
          <button
            type="button"
            role="tab"
            key={label}
            id={`${id}-tab-${i}`}
            aria-controls={`${id}-panel-${i}`}
            aria-selected={active === i}
            tabIndex={active === i ? 0 : -1}
            onClick={() => setActive(i)}
          >
            {label}
          </button>
        ))}
      </div>
      {tabs.map(({ label, content }, i) => (
        <div
          key={label}
          role="tabpanel"
          id={`${id}-panel-${i}`}
          aria-labelledby={`${id}-tab-${i}`}
          hidden={active !== i}
          tabIndex={0}
        >
          {content}
        </div>
      ))}
    </div>
  );
}
