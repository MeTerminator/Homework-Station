import { useSession } from "@/app/providers/session-provider";
import { Empty } from "@/components/feedback/empty";
import { Navigate, Outlet } from "react-router-dom";
export function Guard({ admin = false }: { admin?: boolean }) {
  const { user, loading } = useSession();
  if (loading) return <Empty title="正在载入学习空间…" description="" />;
  if (!user) return <Navigate to="/login" replace />;
  if (admin && !user.admin) return <Navigate to="/" replace />;
  return <Outlet />;
}
