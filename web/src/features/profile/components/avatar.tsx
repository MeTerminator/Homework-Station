import { fileURL } from "@/lib/api/urls";
import type { User } from "@/types/models";
export function Avatar({ user }: { user: User }) {
  return user.avatar_id ? (
    <img
      className="avatar"
      alt={`${user.username} 的头像`}
      src={fileURL(user.avatar_id, true)}
    />
  ) : (
    <span className="avatar">{user.username.slice(0, 1).toUpperCase()}</span>
  );
}
