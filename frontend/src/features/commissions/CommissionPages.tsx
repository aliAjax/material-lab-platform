import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ArrowLeft, CheckCircle2, FilePlus2, Save } from "lucide-react";
import { useForm } from "react-hook-form";
import { Link, useNavigate, useParams } from "react-router-dom";
import { z } from "zod";
import { commissionsApi } from "../../api/resources";
import {
  Button,
  ConfirmDialog,
  ErrorState,
  Field,
  LoadingState,
  PageHeader,
} from "../../components/ui";
import { useUnsavedChanges } from "../../hooks/useUnsavedChanges";

const schema = z.object({
  client_name: z.string().min(1, "请输入委托方名称"),
  material_grade: z.string(),
  batch_no: z.string(),
  sample_description: z.string(),
  test_requirements: z.string(),
  received_at: z.string(),
});
type Value = z.infer<typeof schema>;
export function CommissionFormPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const client = useQueryClient();
  const [confirming, setConfirming] = React.useState(false);
  const detail = useQuery({
    queryKey: ["commission", id],
    queryFn: () => commissionsApi.get(id!),
    enabled: Boolean(id),
  });
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isDirty },
  } = useForm<Value>({
    resolver: zodResolver(schema),
    defaultValues: {
      client_name: "",
      material_grade: "",
      batch_no: "",
      sample_description: "",
      test_requirements: "",
      received_at: new Date().toISOString().slice(0, 16),
    },
  });
  React.useEffect(() => {
    if (detail.data)
      reset({
        ...detail.data,
        received_at: detail.data.received_at.slice(0, 16),
      });
  }, [detail.data, reset]);
  useUnsavedChanges(isDirty);
  const save = useMutation({
    mutationFn: (value: Value) => commissionsApi.save({ ...value, id }),
    onSuccess: (value) => {
      client.invalidateQueries({ queryKey: ["commissions"] });
      reset(value);
      if (!id) navigate(`/commissions/${value.id}/edit`, { replace: true });
    },
  });
  const submit = useMutation({
    mutationFn: async () => {
      const value = await save.mutateAsync(
        (document.querySelector("#commission-form") as HTMLFormElement) &&
          (await new Promise<Value>((resolve, reject) =>
            handleSubmit(resolve, reject)(),
          )),
      );
      return commissionsApi.submit(value.id);
    },
    onSuccess: (value) => {
      setConfirming(false);
      reset(value);
      navigate(`/samples/${value.id}`);
    },
  });
  if (id && detail.isPending) return <LoadingState />;
  if (detail.isError) return <ErrorState error={detail.error} />;
  return (
    <>
      <PageHeader
        title={
          id ? `编辑委托 ${detail.data?.commission_no || ""}` : "新建送检委托"
        }
        description="草稿可暂缺信息，正式提交后将生成不可重复的样品编号。"
        actions={
          <Link className="button button-secondary" to="/samples">
            <ArrowLeft size={16} />
            返回列表
          </Link>
        }
      />
      <form
        id="commission-form"
        className="form-sheet"
        onSubmit={handleSubmit((value) => save.mutate(value))}
      >
        <div className="form-grid">
          <Field
            label="委托方名称"
            required
            error={errors.client_name?.message}
          >
            <input {...register("client_name")} />
          </Field>
          <Field label="接收时间">
            <input type="datetime-local" {...register("received_at")} />
          </Field>
          <Field label="材料牌号">
            <input {...register("material_grade")} />
          </Field>
          <Field label="批号">
            <input {...register("batch_no")} />
          </Field>
          <Field label="样品描述">
            <textarea rows={3} {...register("sample_description")} />
          </Field>
          <Field label="检测要求">
            <textarea rows={3} {...register("test_requirements")} />
          </Field>
        </div>
        {save.error && <div className="form-alert">{save.error.message}</div>}
        <footer className="form-footer">
          <Button
            variant="secondary"
            type="submit"
            disabled={save.isPending}
            icon={<Save size={16} />}
          >
            保存草稿
          </Button>
          {id && (
            <Button
              type="button"
              onClick={() => setConfirming(true)}
              icon={<CheckCircle2 size={16} />}
            >
              提交并生成样品
            </Button>
          )}
        </footer>
      </form>
      <ConfirmDialog
        open={confirming}
        title="正式提交委托？"
        description="提交后会执行完整校验并生成样品编号，关键身份信息将进入审计记录。"
        confirmLabel="确认提交"
        onClose={() => setConfirming(false)}
        onConfirm={() => submit.mutate()}
      />
    </>
  );
}

import React from "react";
export function NewCommissionShortcut() {
  return (
    <Link className="button button-primary" to="/commissions/new">
      <FilePlus2 size={16} />
      新建委托
    </Link>
  );
}
