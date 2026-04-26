import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { beforeEach, describe, expect, test, vi } from 'vitest';
import ServiceVitalsPage from './ServiceVitalsPage';
import { fetchServiceVitals, type ServiceVitals } from '../../../services/project';

vi.mock('../../../services/project', async () => {
  const actual = await vi.importActual<typeof import('../../../services/project')>('../../../services/project');
  return {
    ...actual,
    fetchServiceVitals: vi.fn(),
  };
});

function renderPage() {
  return render(
    <MemoryRouter initialEntries={['/projects/proj-1/services/svc-1']}>
      <Routes>
        <Route path="/projects/:id/services/:serviceId" element={<ServiceVitalsPage />} />
      </Routes>
    </MemoryRouter>,
  );
}

const baseVitals: ServiceVitals = {
  service: {
    id: 'svc-1',
    projectId: 'proj-1',
    name: 'api-service',
    description: '',
    repoUrl: 'https://example.test/repo.git',
    serviceType: 'api',
    language: 'go',
    owner: 'platform',
    tags: ['critical'],
    pipelineIds: ['pipe-1'],
    environmentIds: ['env-1'],
    status: 'active',
    createdAt: '2026-04-26T00:00:00Z',
    updatedAt: '2026-04-26T00:00:00Z',
  },
  summary: {
    status: 'warning',
    reason: 'Recent pipeline steps need attention',
    lastSignalAt: '2026-04-26T00:01:00Z',
    hasDeliveryData: true,
    hasDeploymentData: true,
    activeWarningCount: 1,
  },
  linkedPipelines: [{
    id: 'pipe-1',
    name: 'main-ci',
    status: 'active',
    repoUrl: 'https://example.test/repo.git',
    createdAt: '2026-04-26T00:00:00Z',
    updatedAt: '2026-04-26T00:00:00Z',
  }],
  recentRuns: [{
    id: 'run-1',
    pipelineId: 'pipe-1',
    pipelineName: 'main-ci',
    runNumber: 7,
    status: 'failed',
    errorMessage: 'unit failed',
    createdAt: '2026-04-26T00:02:00Z',
  }],
  latestDeployments: [{
    id: 'dep-1',
    environmentId: 'env-1',
    environmentName: 'prod',
    image: 'repo/app:1',
    status: 'healthy',
    createdAt: '2026-04-26T00:03:00Z',
  }],
  activeSignals: [{
    id: 'sig-1',
    targetType: 'integration',
    targetId: 'registry',
    source: 'registry-test',
    status: 'degraded',
    rawStatus: 'degraded',
    severity: 'critical',
    reason: 'registry.http_error',
    message: 'Registry API returned 401 Unauthorized',
    observedAt: '2026-04-26T00:04:00Z',
  }],
  warnings: [{
    stepName: 'unit-test',
    taskRunName: 'build',
    status: 'failed',
    runId: 'run-1',
    pipelineId: 'pipe-1',
    pipelineName: 'main-ci',
    runNumber: 7,
    runPath: '/projects/proj-1/pipelines/pipe-1/runs/run-1',
    createdAt: '2026-04-26T00:05:00Z',
  }],
  emptyStates: [],
  refreshedAt: '2026-04-26T00:06:00Z',
};

describe('ServiceVitalsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  test('renders populated service vitals evidence', async () => {
    vi.mocked(fetchServiceVitals).mockResolvedValue(baseVitals);

    renderPage();

    expect(await screen.findByRole('heading', { name: 'api-service' })).toBeInTheDocument();
    expect(fetchServiceVitals).toHaveBeenCalledWith('proj-1', 'svc-1');
    expect(screen.getAllByText('Recent pipeline steps need attention').length).toBeGreaterThan(0);
    expect(screen.getAllByText('main-ci').length).toBeGreaterThan(0);
    expect(screen.getByText('build / unit-test')).toBeInTheDocument();
    expect(screen.getByText('prod')).toBeInTheDocument();
    expect(screen.getByText('integration:registry')).toBeInTheDocument();
    expect(screen.getByText('Registry API returned 401 Unauthorized')).toBeInTheDocument();
  });

  test('renders missing evidence and empty states', async () => {
    vi.mocked(fetchServiceVitals).mockResolvedValue({
      ...baseVitals,
      summary: {
        ...baseVitals.summary,
        status: 'unknown',
        reason: 'No delivery evidence yet',
        hasDeliveryData: false,
        hasDeploymentData: false,
        activeWarningCount: 0,
      },
      linkedPipelines: [],
      recentRuns: [],
      latestDeployments: [],
      activeSignals: [],
      warnings: [],
      emptyStates: ['No linked pipelines', 'No deployments'],
    });

    renderPage();

    expect(await screen.findByText('Missing evidence')).toBeInTheDocument();
    expect(screen.getAllByText('No linked pipelines').length).toBeGreaterThan(0);
    expect(screen.getAllByText('No deployments').length).toBeGreaterThan(0);
    expect(screen.getByText('No recent runs')).toBeInTheDocument();
    expect(screen.getByText('No health signals')).toBeInTheDocument();
  });

  test('renders a recoverable error state when vitals fail to load', async () => {
    vi.mocked(fetchServiceVitals).mockRejectedValue(new Error('backend down'));

    renderPage();

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('加载失败：backend down');
    });
    expect(screen.queryByText('加载中...')).not.toBeInTheDocument();
  });
});
