import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { Navigate, Route, Routes, useLocation } from "react-router-dom";
import { AppShell } from "../components/AppShell";
import { LoadingState } from "../components/ui";
import {
  CertificateDetailPage,
  CertificatesPage,
  PublicVerificationPage,
} from "../features/certificates/CertificatePages";
import { CommissionFormPage } from "../features/commissions/CommissionPages";
import { CustodyPage } from "../features/custody/CustodyPage";
import { DashboardPage } from "../features/dashboard/DashboardPage";
import { LoginPage } from "../features/auth/LoginPage";
import { MethodEditorPage, MethodsPage } from "../features/methods/MethodPages";
import { ProfilePage } from "../features/profile/ProfilePage";
import { ReviewDetailPage, ReviewsPage } from "../features/reviews/ReviewPages";
import {
  SampleDetailPage,
  SampleLabelPage,
  SamplesPage,
} from "../features/samples/SamplePages";
import { TaskExecutionPage, TasksPage } from "../features/tasks/TaskPages";
import { AuthProvider, useAuth } from "../state/AuthContext";

const client = new QueryClient({
  defaultOptions: {
    queries: { staleTime: 20_000, retry: 1, refetchOnWindowFocus: false },
  },
});
function Protected() {
  const { user, loading } = useAuth();
  const location = useLocation();
  if (loading) return <LoadingState label="正在恢复会话" />;
  return user ? (
    <AppShell />
  ) : (
    <Navigate
      to="/login"
      state={{ from: location.pathname + location.search }}
      replace
    />
  );
}
export function App() {
  return (
    <QueryClientProvider client={client}>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/verify" element={<PublicVerificationPage />} />
          <Route element={<Protected />}>
            <Route index element={<DashboardPage />} />
            <Route path="samples" element={<SamplesPage />} />
            <Route path="samples/:id" element={<SampleDetailPage />} />
            <Route path="samples/:id/label" element={<SampleLabelPage />} />
            <Route path="commissions/new" element={<CommissionFormPage />} />
            <Route
              path="commissions/:id/edit"
              element={<CommissionFormPage />}
            />
            <Route path="custody" element={<CustodyPage />} />
            <Route path="tasks" element={<TasksPage />} />
            <Route path="tasks/:id" element={<TaskExecutionPage />} />
            <Route path="reviews" element={<ReviewsPage />} />
            <Route path="reviews/:id" element={<ReviewDetailPage />} />
            <Route path="certificates" element={<CertificatesPage />} />
            <Route
              path="certificates/:id"
              element={<CertificateDetailPage />}
            />
            <Route path="methods" element={<MethodsPage />} />
            <Route path="methods/new" element={<MethodEditorPage />} />
            <Route path="methods/:id" element={<MethodEditorPage />} />
            <Route path="profile" element={<ProfilePage />} />
          </Route>
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </AuthProvider>
    </QueryClientProvider>
  );
}
