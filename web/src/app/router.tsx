import { Empty } from "@/components/feedback/empty";
import { Shell } from "@/components/layout/app-shell";
import { lazy, Suspense } from "react";
import { Navigate, Route, Routes } from "react-router-dom";
import { Toaster } from "sonner";
import { useTheme } from "./providers/theme-provider";
import { Guard } from "./route-guard";
const AuthPage = lazy(() =>
  import("@/features/auth/pages/login-page").then((m) => ({ default: m.AuthPage })),
);
const MyAssignments = lazy(() =>
  import("@/features/assignments/pages/my-assignments-page").then((m) => ({
    default: m.MyAssignments,
  })),
);
const AssignmentDetail = lazy(() =>
  import("@/features/assignments/pages/assignment-detail-page").then((m) => ({
    default: m.AssignmentDetail,
  })),
);
const MyAttachments = lazy(() =>
  import("@/features/attachments/pages/my-attachments-page").then((m) => ({
    default: m.MyAttachments,
  })),
);
const Profile = lazy(() =>
  import("@/features/profile/pages/profile-page").then((m) => ({ default: m.Profile })),
);
const Overview = lazy(() =>
  import("@/features/admin/overview/pages/overview-page").then((m) => ({
    default: m.Overview,
  })),
);
const AssignmentList = lazy(() =>
  import("@/features/admin/assignments/pages/assignment-list-page").then((m) => ({
    default: m.AssignmentList,
  })),
);
const AssignmentEditor = lazy(() =>
  import("@/features/admin/assignments/pages/assignment-editor-page").then((m) => ({
    default: m.AssignmentEditor,
  })),
);
const CategoryManager = lazy(() =>
  import("@/features/admin/categories/pages/category-manager-page").then((m) => ({
    default: m.CategoryManager,
  })),
);
const SettingsPage = lazy(() =>
  import("@/features/admin/settings/pages/settings-page").then((m) => ({
    default: m.SettingsPage,
  })),
);
const UserList = lazy(() =>
  import("@/features/admin/users/pages/user-list-page").then((m) => ({
    default: m.UserList,
  })),
);
const UserEditor = lazy(() =>
  import("@/features/admin/users/pages/user-editor-page").then((m) => ({
    default: m.UserEditor,
  })),
);
const AttachmentManager = lazy(() =>
  import("@/features/admin/attachments/pages/attachment-manager-page").then((m) => ({
    default: m.AttachmentManager,
  })),
);
export function AppRouter() {
  const { dark } = useTheme();
  return (
    <>
      <Suspense fallback={<Empty title="正在加载页面…" description="" />}>
        <Routes>
          <Route path="/login" element={<AuthPage />} />
          <Route element={<Guard />}>
            <Route element={<Shell />}>
              <Route index element={<MyAssignments />} />
              <Route path="profile" element={<Profile />} />
              <Route path="attachments" element={<MyAttachments />} />
              <Route path="assignments/:id" element={<AssignmentDetail />} />
            </Route>
          </Route>
          <Route element={<Guard admin />}>
            <Route path="/admin" element={<Shell admin />}>
              <Route index element={<Overview />} />
              <Route path="assignments" element={<AssignmentList />} />
              <Route path="assignments/:id" element={<AssignmentEditor />} />
              <Route path="categories" element={<CategoryManager />} />
              <Route path="users" element={<UserList />} />
              <Route path="users/:id" element={<UserEditor />} />
              <Route path="attachments" element={<AttachmentManager />} />
              <Route path="settings" element={<SettingsPage />} />
            </Route>
          </Route>
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </Suspense>
      <Toaster theme={dark ? "dark" : "light"} richColors position="top-right" />
    </>
  );
}
