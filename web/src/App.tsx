import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';
import { ErrorBoundary } from './components/common/ErrorBoundary';
import { RequireAuth } from './components/common/RequireAuth';
import { RequirePermission } from './components/common/RequirePermission';
import { AdminUsersPage } from './pages/admin-users/AdminUsersPage';
import { DashboardPage } from './pages/dashboard/DashboardPage';
import { ForbiddenPage } from './pages/forbidden/ForbiddenPage';
import { LoginPage } from './pages/login/LoginPage';
import { ProjectListPage } from './pages/projects/ProjectListPage';
import { ProjectLayout } from './pages/projects/ProjectLayout';
import { EnvironmentListPage } from './pages/projects/environments/EnvironmentListPage';
import { DeploymentListPage } from './pages/projects/deployments/DeploymentListPage';
import { DeploymentDetailPage } from './pages/projects/deployments/DeploymentDetailPage';
import { ServiceListPage } from './pages/projects/services/ServiceListPage';
import { MemberListPage } from './pages/projects/members/MemberListPage';
import { VariableListPage } from './pages/projects/variables/VariableListPage';
import { AdminVariablePage } from './pages/admin/variables/AdminVariablePage';
import { IntegrationsPage } from './pages/admin/integrations/IntegrationsPage';
import { lazy, Suspense } from 'react';
import { NotFoundPage } from './pages/notfound/NotFoundPage';
import { PageSkeleton } from './components/common/PageSkeleton';

const PipelineListPage = lazy(() => import('./pages/projects/pipelines/PipelineListPage'));
const PipelineEditorPage = lazy(() => import('./pages/projects/pipelines/PipelineEditorPage'));
const PipelineRunListPage = lazy(() => import('./pages/projects/pipelines/PipelineRunListPage'));
const PipelineRunDetailPage = lazy(() => import('./pages/projects/pipelines/PipelineRunDetailPage'));
const TemplateSelectPage = lazy(() => import('./pages/projects/pipelines/TemplateSelectPage'));
const AnalyticsPage = lazy(() => import('./pages/projects/analytics/AnalyticsPage'));
const ServiceVitalsPage = lazy(() => import('./pages/projects/services/ServiceVitalsPage'));
const NotificationRulesPage = lazy(() => import('./pages/projects/notifications/NotificationRulesPage'));
const AuditLogPage = lazy(() => import('./pages/admin/audit/AuditLogPage'));
const AccessTokensPage = lazy(() => import('./pages/access-tokens/AccessTokensPage'));
const SystemSettingsPage = lazy(() => import('./pages/admin/settings/SystemSettingsPage'));

function App() {
  return (
    <ErrorBoundary>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route
            path="/dashboard"
            element={
              <RequireAuth>
                <DashboardPage />
              </RequireAuth>
            }
          />
          <Route
            path="/admin/users"
            element={
              <RequireAuth>
                <RequirePermission permission="route:admin-users:view">
                  <AdminUsersPage />
                </RequirePermission>
              </RequireAuth>
            }
          />
          <Route
            path="/projects"
            element={
              <RequireAuth>
                <ProjectListPage />
              </RequireAuth>
            }
          />
          <Route
            path="/projects/:id"
            element={
              <RequireAuth>
                <ProjectLayout />
              </RequireAuth>
            }
          >
            <Route index element={<Navigate to="environments" replace />} />
            <Route path="environments" element={<EnvironmentListPage />} />
            <Route path="deployments" element={<DeploymentListPage />} />
            <Route path="deployments/:deployId" element={<DeploymentDetailPage />} />
            <Route path="services" element={<ServiceListPage />} />
            <Route path="services/:serviceId" element={<Suspense fallback={<PageSkeleton />}><ServiceVitalsPage /></Suspense>} />
            <Route path="members" element={<MemberListPage />} />
            <Route path="variables" element={<VariableListPage />} />
            <Route path="pipelines" element={<Suspense fallback={<PageSkeleton />}><PipelineListPage /></Suspense>} />
            <Route path="analytics" element={<Suspense fallback={<PageSkeleton />}><AnalyticsPage /></Suspense>} />
            <Route path="pipelines/new" element={<Suspense fallback={<PageSkeleton />}><TemplateSelectPage /></Suspense>} />
            <Route path="pipelines/:pipelineId/runs" element={<Suspense fallback={<PageSkeleton />}><PipelineRunListPage /></Suspense>} />
            <Route path="pipelines/:pipelineId/runs/:runId" element={<Suspense fallback={<PageSkeleton />}><PipelineRunDetailPage /></Suspense>} />
            <Route path="notifications" element={<Suspense fallback={<PageSkeleton />}><NotificationRulesPage /></Suspense>} />
          </Route>
          {/* Fullscreen pipeline editor — outside ProjectLayout for maximum canvas space */}
          <Route
            path="/projects/:id/pipelines/blank"
            element={
              <RequireAuth>
                <Suspense fallback={<PageSkeleton />}><PipelineEditorPage /></Suspense>
              </RequireAuth>
            }
          />
          <Route
            path="/projects/:id/pipelines/:pipelineId"
            element={
              <RequireAuth>
                <Suspense fallback={<PageSkeleton />}><PipelineEditorPage /></Suspense>
              </RequireAuth>
            }
          />
          <Route
            path="/admin/variables"
            element={
              <RequireAuth>
                <RequirePermission permission="route:admin-variables:view">
                  <AdminVariablePage />
                </RequirePermission>
              </RequireAuth>
            }
          />
          <Route
            path="/admin/integrations"
            element={
              <RequireAuth>
                <RequirePermission permission="route:admin-integrations:view">
                  <IntegrationsPage />
                </RequirePermission>
              </RequireAuth>
            }
          />
          <Route
            path="/admin/audit-logs"
            element={
              <RequireAuth>
                <RequirePermission permission="route:admin-audit:view">
                  <Suspense fallback={<PageSkeleton />}><AuditLogPage /></Suspense>
                </RequirePermission>
              </RequireAuth>
            }
          />
          <Route
            path="/admin/access-tokens"
            element={
              <RequireAuth>
                <RequirePermission permission="route:access-tokens:view">
                  <Suspense fallback={<PageSkeleton />}><AccessTokensPage /></Suspense>
                </RequirePermission>
              </RequireAuth>
            }
          />
          <Route
            path="/admin/settings"
            element={
              <RequireAuth>
                <RequirePermission permission="route:admin-settings:view">
                  <Suspense fallback={<PageSkeleton />}><SystemSettingsPage /></Suspense>
                </RequirePermission>
              </RequireAuth>
            }
          />
          <Route
            path="/403"
            element={
              <RequireAuth>
                <ForbiddenPage />
              </RequireAuth>
            }
          />
          <Route path="/" element={<Navigate to="/dashboard" replace />} />
          <Route path="*" element={<RequireAuth><NotFoundPage /></RequireAuth>} />
        </Routes>
      </BrowserRouter>
    </ErrorBoundary>
  );
}

export default App;
