export async function api<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers: Record<string, string> = {};
  if (
    options.body &&
    !(options.body instanceof FormData) &&
    !(options.body instanceof Blob)
  ) {
    headers["Content-Type"] = "application/json";
  }
  const res = await fetch("/api" + path, {
    ...options,
    headers: { ...headers, ...options.headers },
    credentials: "same-origin",
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "网络请求失败" }));
    if (res.status === 401 && path !== "/login" && path !== "/me") {
      window.dispatchEvent(new Event("session-expired"));
    }
    throw new Error(err.error || `请求失败 (${res.status})`);
  }
  return res.json();
}

export const json = (method: string, data?: unknown): RequestInit => ({
  method,
  body: data === undefined ? undefined : JSON.stringify(data),
});
