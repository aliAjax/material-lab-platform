import { useMutation, useQuery } from "@tanstack/react-query";
import { Ban, FileCheck2, Printer, Search, ShieldCheck } from "lucide-react";
import { useState } from "react";
import { Link, useParams } from "react-router-dom";
import { certificatesApi } from "../../api/resources";
import {
  Button,
  ConfirmDialog,
  EmptyState,
  ErrorState,
  Field,
  LoadingState,
  PageHeader,
  StatusBadge,
  Toast,
} from "../../components/ui";

export function CertificatesPage() {
  const query = useQuery({
    queryKey: ["certificates"],
    queryFn: certificatesApi.list,
  });
  const [commission, setCommission] = useState("");
  const [open, setOpen] = useState(false);
  const issue = useMutation({
    mutationFn: () => certificatesApi.issue(commission),
    onSuccess: () => {
      setOpen(false);
      setCommission("");
      void query.refetch();
    },
  });
  return (
    <>
      <PageHeader
        title="检验证书"
        description="所有必需任务通过复核后，由实验室负责人签发。"
        actions={
          <Button onClick={() => setOpen(true)} icon={<FileCheck2 size={16} />}>
            签发证书
          </Button>
        }
      />
      {open && (
        <section className="inline-form">
          <h2>签发确认</h2>
          <Field label="委托单 ID" required>
            <input
              value={commission}
              onChange={(e) => setCommission(e.target.value)}
            />
          </Field>
          <p className="muted">
            服务端将再次验证任务完整性，生成唯一编号和 SHA-256 摘要。
          </p>
          <footer>
            <Button variant="secondary" onClick={() => setOpen(false)}>
              取消
            </Button>
            <Button
              disabled={!commission || issue.isPending}
              onClick={() => issue.mutate()}
            >
              确认签发
            </Button>
          </footer>
        </section>
      )}
      {query.isPending ? (
        <LoadingState />
      ) : query.isError ? (
        <ErrorState error={query.error} />
      ) : !query.data!.items.length ? (
        <EmptyState />
      ) : (
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>证书编号</th>
                <th>样品</th>
                <th>版本</th>
                <th>签发时间</th>
                <th>状态</th>
              </tr>
            </thead>
            <tbody>
              {query.data!.items.map((item) => (
                <tr key={item.id}>
                  <td>
                    <Link
                      className="strong-link"
                      to={`/certificates/${item.id}`}
                    >
                      {item.certificate_no}
                    </Link>
                  </td>
                  <td>{item.sample_no}</td>
                  <td>v{item.version}</td>
                  <td>{new Date(item.issued_at).toLocaleString()}</td>
                  <td>
                    <StatusBadge
                      value={item.status}
                      tone={item.status === "valid" ? "success" : "danger"}
                    />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </>
  );
}
export function CertificateDetailPage() {
  const { id } = useParams();
  const query = useQuery({
    queryKey: ["certificate", id],
    queryFn: () => certificatesApi.get(id!),
  });
  const [voiding, setVoiding] = useState(false);
  const [reason, setReason] = useState("");
  const [message, setMessage] = useState("");
  const voidMutation = useMutation({
    mutationFn: () => certificatesApi.void(id!, reason),
    onSuccess: () => {
      setVoiding(false);
      setMessage("原证书已作废，可签发关联更正版");
      void query.refetch();
    },
  });
  if (query.isPending) return <LoadingState />;
  if (query.isError) return <ErrorState error={query.error} />;
  const value = query.data!;
  return (
    <>
      <PageHeader
        title={value.certificate_no}
        description={`证书版本 v${value.version}`}
        actions={
          <>
            <Button
              variant="secondary"
              onClick={() => window.print()}
              icon={<Printer size={16} />}
            >
              打印
            </Button>
            {value.status === "valid" && (
              <Button
                variant="danger"
                onClick={() => setVoiding(true)}
                icon={<Ban size={16} />}
              >
                作废
              </Button>
            )}
          </>
        }
      />
      <article className="certificate">
        <header>
          <ShieldCheck size={34} />
          <div>
            <span>材检实验室</span>
            <h2>工业材料检验证书</h2>
          </div>
          <StatusBadge
            value={value.status}
            tone={value.status === "valid" ? "success" : "danger"}
          />
        </header>
        <dl>
          <dt>证书编号</dt>
          <dd>{value.certificate_no}</dd>
          <dt>样品编号</dt>
          <dd>{value.sample_no}</dd>
          <dt>签发时间</dt>
          <dd>{new Date(value.issued_at).toLocaleString()}</dd>
          <dt>检测方法</dt>
          <dd>{value.method_versions?.join("、") || "-"}</dd>
        </dl>
        <table>
          <thead>
            <tr>
              <th>项目</th>
              <th>结果</th>
              <th>判定</th>
            </tr>
          </thead>
          <tbody>
            {value.results?.map((result) => (
              <tr key={result.key}>
                <td>{result.label}</td>
                <td>
                  {result.value} {result.unit}
                </td>
                <td>{result.decision}</td>
              </tr>
            ))}
          </tbody>
        </table>
        <footer>
          <span>SHA-256 摘要</span>
          <code>{value.digest}</code>
        </footer>
      </article>
      {voiding && (
        <section className="inline-form">
          <Field label="作废原因" required>
            <textarea
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              rows={3}
            />
          </Field>
        </section>
      )}
      <ConfirmDialog
        open={voiding}
        title="作废这份证书？"
        description="证书不可覆盖，作废后状态永久保留；更正内容必须签发关联的新版本。"
        danger
        confirmLabel="确认作废"
        onClose={() => setVoiding(false)}
        onConfirm={() =>
          reason.trim() ? voidMutation.mutate() : setMessage("必须填写作废原因")
        }
      />
      {message && (
        <Toast
          message={message}
          tone={message.includes("必须") ? "danger" : "success"}
          onClose={() => setMessage("")}
        />
      )}
    </>
  );
}
export function PublicVerificationPage() {
  const [code, setCode] = useState("");
  const [submitted, setSubmitted] = useState("");
  const query = useQuery({
    queryKey: ["verify-certificate", submitted],
    queryFn: () => certificatesApi.verify(submitted),
    enabled: Boolean(submitted),
    retry: false,
  });
  return (
    <main className="verify-page">
      <section className="verify-header">
        <ShieldCheck size={30} />
        <div>
          <h1>证书公开核验</h1>
          <p>输入证书上的高熵校验码</p>
        </div>
      </section>
      <form
        onSubmit={(e) => {
          e.preventDefault();
          setSubmitted(code.trim());
        }}
      >
        <Field label="校验码">
          <input
            autoFocus
            value={code}
            onChange={(e) => setCode(e.target.value)}
            autoComplete="off"
          />
        </Field>
        <Button disabled={!code.trim()} icon={<Search size={16} />}>
          查询
        </Button>
      </form>
      {query.isLoading ? (
        <LoadingState />
      ) : query.isError ? (
        <div className="verification-result invalid">
          <Ban />
          <div>
            <strong>未找到有效证书</strong>
            <p>{query.error.message}</p>
          </div>
        </div>
      ) : query.data ? (
        <div className="verification-result valid">
          <ShieldCheck />
          <div>
            <strong>证书核验通过</strong>
            <dl>
              <dt>证书编号</dt>
              <dd>{query.data.certificate_no}</dd>
              <dt>样品编号</dt>
              <dd>{query.data.sample_no}</dd>
              <dt>状态</dt>
              <dd>{query.data.status}</dd>
              <dt>签发时间</dt>
              <dd>{new Date(query.data.issued_at).toLocaleString()}</dd>
            </dl>
          </div>
        </div>
      ) : null}
      <footer>
        <Link to="/login">实验室人员登录</Link>
      </footer>
    </main>
  );
}
