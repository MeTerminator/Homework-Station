import { Inbox } from "lucide-react";
import type { ReactNode } from "react";

export function Empty({
  title = "暂无数据",
  description = "新的内容会出现在这里。",
  children,
}: {
  title?: string;
  description?: string;
  children?: ReactNode;
}) {
  return (
    <div className="empty">
      <Inbox size={32} />
      <h3>{title}</h3>
      <p className="small" style={{ marginBottom: children ? 18 : 0 }}>
        {description}
      </p>
      {children}
    </div>
  );
}
