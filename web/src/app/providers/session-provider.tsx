import { DEFAULT_QQ_BINDING_INSTRUCTIONS } from "@/lib/qq-binding";
import { api } from "@/lib/api/client";
import type { User } from "@/types/models";
import type { ReactNode } from "react";
import { createContext, useContext, useEffect, useState } from "react";
export interface Session {
  user: User | null;
  setUser: (u: User | null) => void;
  loading: boolean;
  title: string;
  logo: string;
  qqBindingInstructions: string;
  refreshTitle: () => Promise<void>;
}

export const SessionContext = createContext<Session>(null!);

export const useSession = () => useContext(SessionContext);

export function SessionProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [title, setTitle] = useState("作业驿站");
  const [logo, setLogo] = useState("");
  const [qqBindingInstructions, setQQBindingInstructions] = useState(
    DEFAULT_QQ_BINDING_INSTRUCTIONS,
  );
  const refreshTitle = () => {
    return api<{
      site_title: string;
      browser_title: string;
      site_logo: string;
      site_icon: string;
      qq_binding_instructions: string;
    }>("/public-settings")
      .then((v) => {
        setTitle(v.site_title || "作业驿站");
        setLogo(v.site_logo || "");
        setQQBindingInstructions(
          v.qq_binding_instructions || DEFAULT_QQ_BINDING_INSTRUCTIONS,
        );
        document.title = v.browser_title || v.site_title || "作业驿站";
        let icon = document.querySelector<HTMLLinkElement>('link[rel="icon"]');
        if (v.site_icon) {
          if (!icon) {
            icon = document.createElement("link");
            icon.rel = "icon";
            document.head.appendChild(icon);
          }
          icon.href = v.site_icon;
        } else {
          icon?.remove();
        }
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
    <SessionContext.Provider
      value={{ user, setUser, loading, title, logo, qqBindingInstructions, refreshTitle }}
    >
      {children}
    </SessionContext.Provider>
  );
}
