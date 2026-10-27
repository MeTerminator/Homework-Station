import { Button } from "@/components/ui/button";
import { Loader2 } from "lucide-react";
import type { ReactNode } from "react";
export function LoadState({
  loading,
  error,
  retry,
  children,
}: {
  loading: boolean;
  error: string;
  retry?: () => void;
  children: ReactNode;
}) {
  if (loading)
    return (
      <div className="empty row" style={{ justifyContent: "center" }}>
        <Loader2 size={20} className="animate-spin" />
        正在加载…
      </div>
    );
  if (error)
    return (
      <div className="error-box row">
        {error}
        <Button variant="outline" size="sm" onClick={retry}>
          重试
        </Button>
      </div>
    );
  return <>{children}</>;
}
