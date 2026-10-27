import { api, json } from "@/lib/api/client";
import type { Attachment } from "@/types/models";
import { createSHA256 } from "hash-wasm";
export async function upload(
  file: File,
  assignmentID: number | undefined,
  progress: (p: number, label: string) => void,
  signal: AbortSignal,
): Promise<Attachment> {
  const hash = await createSHA256();
  hash.init();
  for (let offset = 0; offset < file.size; offset += 4 * 1024 * 1024) {
    signal.throwIfAborted();
    hash.update(
      new Uint8Array(await file.slice(offset, offset + 4 * 1024 * 1024).arrayBuffer()),
    );
    progress(file.size ? (offset / file.size) * 10 : 10, "正在计算 SHA256");
  }
  let a = await api<Attachment>("/uploads", {
    ...json("POST", {
      name: file.name,
      size: file.size,
      sha256: hash.digest(),
      assignment_id: assignmentID,
    }),
    signal,
  });
  if (a.status === "complete") {
    progress(100, "已复用附件");
    return a;
  }
  const done = new Set(a.received || []);
  for (let i = 0; i < a.total_chunks; i++) {
    signal.throwIfAborted();
    if (!done.has(i)) {
      let last: unknown;
      for (let attempt = 0; attempt < 3; attempt++) {
        try {
          await api(`/uploads/${a.id}/chunks/${i}`, {
            method: "PUT",
            body: file.slice(
              i * a.chunk_size,
              Math.min((i + 1) * a.chunk_size, file.size),
            ),
            signal,
          });
          last = null;
          break;
        } catch (e) {
          last = e;
          if (signal.aborted) throw e;
        }
      }
      if (last) throw last;
      done.add(i);
    }
    progress(
      10 + (done.size / Math.max(a.total_chunks, 1)) * 85,
      `上传分片 ${done.size}/${a.total_chunks}`,
    );
  }
  progress(96, "正在合并并校验");
  a = await api<Attachment>(`/uploads/${a.id}/complete`, {
    method: "POST",
    signal,
  });
  progress(100, "上传完成");
  return a;
}
