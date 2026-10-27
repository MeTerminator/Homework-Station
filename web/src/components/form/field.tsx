import type { ReactElement, ReactNode } from "react";
import { cloneElement, isValidElement, useId } from "react";

export function Field({
  label,
  hint,
  children,
  full = false,
}: {
  label: string;
  hint?: string;
  children: ReactNode;
  full?: boolean;
}) {
  const id = useId();
  const child = isValidElement(children)
    ? cloneElement(children as ReactElement<{ id?: string }>, { id })
    : children;
  return (
    <div className={`field ${full ? "full" : ""}`}>
      <label htmlFor={id}>{label}</label>
      {child}
      {hint && <span className="hint">{hint}</span>}
    </div>
  );
}
