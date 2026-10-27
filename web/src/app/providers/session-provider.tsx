import { api } from "@/lib/api/client";
import type { User } from "@/types/models";
import type { ReactNode } from "react";
import { createContext, useContext, useEffect, useState } from "react";
export interface Session {
  user: User | null;
  setUser: (u: User | null) => void;
  loading: boolean;
  title: string;
  refreshTitle: () => void;
}

export const SessionContext = createContext<Session>(null!);

export const useSession = () => useContext(SessionContext);

export function SessionProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [title, setTitle] = useState("作业驿站");
  const refreshTitle = () => {
    api<{ site_title: string }>("/public-settings")
      .then((v) => {
        setTitle(v.site_title);
        document.title = v.site_title;
      })
      .catch(() => {});
  };
  useEffect(() => {
    api<User>("/me")
      .then(setUser)
      .catch(() => {})
      .finally(() => setLoading(false));
    refreshTitle();
    const expired = () => setUser(null);
    window.addEventListener("session-expired", expired);
    return () => window.removeEventListener("session-expired", expired);
  }, []);
  return (
    <SessionContext.Provider value={{ user, setUser, loading, title, refreshTitle }}>
      {children}
    </SessionContext.Provider>
  );
}
