import { useCallback, useEffect, useState } from 'react';
import { Message } from '@arco-design/web-react';
import { AppLayout } from '../../components/layout/AppLayout';
import { PageHeader } from '../../components/ui/PageHeader';
import { Card } from '../../components/ui/Card';
import { Btn } from '../../components/ui/Btn';
import { StatusBadge } from '../../components/ui/StatusBadge';
import { ICopy, IKey, IPlus } from '../../components/ui/icons';
import { fetchAccessTokens, revokeAccessToken, type AccessTokenRecord } from '../../services/accessToken';
import { extractErrorMessage } from '../../services/http';
import { TokenFormModal } from './TokenFormModal';

function formatTime(value?: string) {
  if (!value) return '-';
  return new Date(value).toLocaleString();
}

function tokenStatus(token: AccessTokenRecord) {
  if (token.revokedAt) return 'revoked';
  if (new Date(token.expiresAt).getTime() <= Date.now()) return 'token_expired';
  return 'active';
}

const SCOPE_LABELS: Record<string, string> = {
  'pipelines:read': '流水线读取',
  'pipelines:trigger': '流水线触发',
  'deployments:read': '部署读取',
  'deployments:write': '部署写入',
  'variables:read': '变量读取',
  'notifications:read': '通知读取',
  'admin:read': '管理读取',
};

export function AccessTokensPage() {
  const [tokens, setTokens] = useState<AccessTokenRecord[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [createdSecret, setCreatedSecret] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const data = await fetchAccessTokens();
      setTokens(data.items || []);
    } catch (error: unknown) {
      Message.error(extractErrorMessage(error, '加载访问令牌失败'));
      setTokens([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { void load(); }, [load]);

  const handleCreated = (record: AccessTokenRecord, rawToken: string) => {
    setCreatedSecret(rawToken);
    setTokens((prev) => [record, ...prev]);
    Message.success('令牌已创建，请立即复制保存');
  };

  const copySecret = async () => {
    if (!createdSecret) return;
    await navigator.clipboard.writeText(createdSecret);
    Message.success('令牌已复制');
  };

  const revoke = async (token: AccessTokenRecord) => {
    try {
      await revokeAccessToken(token.id);
      Message.success('令牌已撤销');
      await load();
    } catch (error: unknown) {
      Message.error(extractErrorMessage(error, '撤销令牌失败'));
    }
  };

  return (
    <AppLayout>
      <PageHeader
        crumb="System · Access Control"
        title="访问令牌"
        sub="为自动化任务创建有过期时间和明确 scope 的 PAT / 项目令牌。"
        actions={<Btn size="sm" variant="primary" icon={<IPlus size={13} />} onClick={() => setModalVisible(true)}>新建令牌</Btn>}
      />
      <div style={{ padding: 24, display: 'flex', flexDirection: 'column', gap: 14 }}>
        {createdSecret && (
          <Card title="一次性令牌明文" extra={<Btn size="xs" icon={<ICopy size={12} />} onClick={copySecret}>复制</Btn>}>
            <p className="sub" style={{ marginTop: 0 }}>关闭或刷新后将无法再次查看。请存入 CI Secret 或密码管理器。</p>
            <code className="code" style={{ display: 'block', wordBreak: 'break-all', padding: 12 }}>{createdSecret}</code>
          </Card>
        )}

        <Card padding={false}>
          {loading ? (
            <div style={{ padding: '40px 0', textAlign: 'center', color: 'var(--z-400)' }}>加载中...</div>
          ) : (
            <table className="table">
              <thead>
                <tr>
                  <th>名称</th><th>类型</th><th>Scopes</th><th>过期时间</th><th>最近使用</th><th>状态</th><th style={{ textAlign: 'right' }}>操作</th>
                </tr>
              </thead>
              <tbody>
                {tokens.map((token) => (
                  <tr key={token.id}>
                    <td>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                        <IKey size={14} />
                        <div><b style={{ fontWeight: 500 }}>{token.name}</b><div className="sub mono" style={{ fontSize: 11 }}>{token.tokenPrefix}{token.id.slice(0, 8)}…</div></div>
                      </div>
                    </td>
                    <td><StatusBadge status={token.type} /></td>
                    <td>
                      <div style={{ display: 'flex', gap: 4, flexWrap: 'wrap' }}>
                        {token.scopes.map((scope) => (
                          <span key={scope} className="badge badge--grey" title={scope}>{SCOPE_LABELS[scope] ?? scope}</span>
                        ))}
                      </div>
                    </td>
                    <td><span className="sub mono" style={{ fontSize: 11 }}>{formatTime(token.expiresAt)}</span></td>
                    <td><span className="sub mono" style={{ fontSize: 11 }}>{formatTime(token.lastUsedAt)}</span></td>
                    <td><StatusBadge status={tokenStatus(token)} /></td>
                    <td style={{ textAlign: 'right' }}>
                      <Btn size="xs" variant="ghost" disabled={Boolean(token.revokedAt)} onClick={() => revoke(token)}>撤销</Btn>
                    </td>
                  </tr>
                ))}
                {tokens.length === 0 && !loading && (
                  <tr><td colSpan={7} style={{ textAlign: 'center', padding: '40px 0', color: 'var(--z-400)' }}>暂无访问令牌</td></tr>
                )}
              </tbody>
            </table>
          )}
        </Card>
      </div>
      <TokenFormModal visible={modalVisible} onClose={() => setModalVisible(false)} onCreated={handleCreated} />
    </AppLayout>
  );
}

export default AccessTokensPage;
