import type { User } from "@/types/models";
export function selectedCategories(u: User | null): number[] {
  try {
    return JSON.parse(u?.category_ids || "[]");
  } catch {
    return [];
  }
}
