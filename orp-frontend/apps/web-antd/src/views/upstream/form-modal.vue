<script lang="ts" setup>
import type { UpstreamGroup, UpstreamGroupPayload, UpstreamNode } from '#/api';

import { ref, watch } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import {
  AutoComplete,
  Button,
  Input,
  InputNumber,
  message,
  Switch,
  Tag,
  Tooltip,
} from 'ant-design-vue';

import { useVbenForm, z } from '#/adapter/form';
import {
  createUpstreamGroupApi,
  listCentersApi,
  updateUpstreamGroupApi,
} from '#/api';
import { useDirtyGuard } from '#/composables/use-dirty-guard';

defineOptions({ name: 'UpstreamFormModal' });

const emit = defineEmits<{ success: [] }>();

const isEdit = ref(false);
const editId = ref(0);

/* ---------------- Upstream 块指令（upstream {} 上下文） ---------------- */

const UPSTREAM_DIRECTIVE_OPTIONS = [
  { value: 'keepalive', label: 'keepalive (保持连接数，例如 32)' },
  { value: 'keepalive_requests', label: 'keepalive_requests (单连接最大请求数，例如 1000)' },
  { value: 'keepalive_timeout', label: 'keepalive_timeout (空闲超时，例如 60s)' },
  { value: 'keepalive_time', label: 'keepalive_time (连接最大存活时间，例如 1h)' },
  { value: 'hash', label: 'hash (自定义哈希键，例如 $request_uri consistent)' },
  { value: 'ip_hash', label: 'ip_hash (客户端 IP 哈希)' },
  { value: 'least_conn', label: 'least_conn (最少连接)' },
  { value: 'least_time', label: 'least_time (最少时间，例如 header / last_byte)' },
  { value: 'random', label: 'random (随机选择，例如 two least_conn)' },
  { value: 'zone', label: 'zone (共享内存区，例如 upstream_backend 64k)' },
];

const DIRECTIVE_NAME_PATTERN = /^[a-z][a-z0-9_-]{0,63}$/;

const editDirectives = ref<Array<{ name: string; value: string }>>([]);

function addDirective() {
  editDirectives.value.push({ name: '', value: '' });
}

function removeDirective(index: number) {
  editDirectives.value.splice(index, 1);
}

function filterDirectiveOption(input: string, option: { label: string }) {
  return option.label.toLowerCase().includes(input.toLowerCase());
}

function validateDirectives(): null | string {
  const seen = new Set<string>();
  for (const [index, item] of editDirectives.value.entries()) {
    const name = item.name.trim();
    if (!name) {
      return `第 ${index + 1} 条 Upstream 指令名不能为空`;
    }
    if (!DIRECTIVE_NAME_PATTERN.test(name)) {
      return `第 ${index + 1} 条指令名「${name}」不合法：仅支持小写字母、数字、下划线及连字符`;
    }
    if (seen.has(name)) {
      return `Upstream 指令 ${name} 重复定义，请合并或删除重复项`;
    }
    seen.add(name);
  }
  return null;
}

const { markDirty, onBeforeClose, resetDirty } = useDirtyGuard();

const [Form, formApi] = useVbenForm({
  handleValuesChange: () => markDirty(),
  handleSubmit: onSubmit,
  schema: [
    {
      component: 'ApiSelect',
      componentProps: {
        allowClear: false,
        api: async () => {
          const res = await listCentersApi({ page: 1, pageSize: 200 });
          return res.items.map((item) => ({ label: item.name, value: item.id }));
        },
        placeholder: '请选择所属中心',
      },
      fieldName: 'centerId',
      label: '所属中心',
      rules: z.number({ message: '请选择所属中心' }).min(1, { message: '请选择所属中心' }),
    },
    {
      component: 'Input',
      componentProps: { placeholder: '例如：web-frontend' },
      fieldName: 'name',
      help: '字母开头，2-63 位字母数字下划线连字符',
      label: '上游组名称',
      rules: z
        .string()
        .min(1, { message: '请输入上游组名称' })
        .regex(/^[a-zA-Z][a-zA-Z0-9_-]{1,62}$/, {
          message: '必须以字母开头，2-63 位字母数字下划线连字符',
        }),
    },
    {
      component: 'Select',
      componentProps: {
        showSearch: true,

        optionFilterProp: 'label',
        options: [
          { label: '轮询 (round_robin)', value: 'round_robin' },
          { label: '最少连接 (least_conn)', value: 'least_conn' },
          { label: 'IP 哈希 (ip_hash)', value: 'ip_hash' },
          { label: '通用哈希 (hash)', value: 'hash' },
          { label: '一致性哈希 (consistent_hash)', value: 'consistent_hash' },
        ],
        placeholder: '默认轮询',
      },
      defaultValue: 'round_robin',
      fieldName: 'lbPolicy',
      label: '负载均衡策略',
    },
    {
      component: 'Input',
      componentProps: { placeholder: '$remote_addr' },
      fieldName: 'hashKey',
      label: 'Hash 键',
      help: '选择 Hash 策略时必填；变量须同时可用于 HTTP 与 Stream，例如 $remote_addr。',
    },
    {
      component: 'Input',
      componentProps: { maxLength: 100, placeholder: '服务用途备注（可选）' },
      fieldName: 'description',
      label: '描述',
    },
    {
      component: 'Input',
      componentProps: { placeholder: '逗号分隔，例如：web,prod' },
      fieldName: 'tagsText',
      label: '标签',
    },
  ],
  showDefaultActions: false,
});

/* ---------------- 后端节点列表（受控动态编辑） ---------------- */

interface EditableNode {
  backup: boolean;
  failTimeoutSec: number;
  host: string;
  maxFails: number;
  port: number | undefined;
  slowStartSec: number;
  weight: number;
}

const editNodes = ref<EditableNode[]>([]);

function addNode() {
  editNodes.value.push({
    backup: false,
    failTimeoutSec: 10,
    host: '',
    maxFails: 3,
    port: undefined,
    slowStartSec: 0,
    weight: 1,
  });
}

function removeNode(index: number) {
  editNodes.value.splice(index, 1);
}

/* ---------------- 健康检查（受控） ---------------- */

const hcType = ref<'http' | 'none' | 'tcp'>('none');
const hcIntervalSec = ref(5);
const hcPath = ref('/healthz');
const hcExpectedStatus = ref('200');

/* ---------------- 弹窗 ---------------- */

watch(editNodes, () => markDirty(), { deep: true });
watch(editDirectives, () => markDirty(), { deep: true });
watch([hcType, hcIntervalSec, hcPath, hcExpectedStatus], () => markDirty());

const [Modal, modalApi] = useVbenModal<{ record?: UpstreamGroup }>({
  fullscreenButton: false,
  onBeforeClose,
  async onConfirm() {
    await formApi.validateAndSubmitForm();
  },
  async onOpenChange(isOpen: boolean) {
    if (isOpen) {
      await formApi.resetForm();
      const record = modalApi.getData()?.record;
      isEdit.value = Boolean(record);
      editId.value = record?.id ?? 0;
      modalApi.setState({ title: record ? '编辑上游组' : '新增上游组' });
      if (record) {
        formApi.setValues({
          centerId: record.centerId,
          description: record.description,
          hashKey: record.hashKey ?? '$remote_addr',
          lbPolicy: record.lbPolicy,
          name: record.name,
          tagsText: record.tags.join(','),
        });
        editNodes.value = record.nodes.map((item) => ({ ...item }));
        editDirectives.value = (record.directives ?? []).map((item) => ({
          name: item.name,
          value: item.value,
        }));
        hcType.value = record.healthCheck.type;
        hcIntervalSec.value = record.healthCheck.intervalSec;
        hcPath.value = record.healthCheck.path || '/healthz';
        hcExpectedStatus.value =
          record.healthCheck.expectedStatus.join(',') || '200';
      } else {
        editNodes.value = [];
        formApi.setValues({ centerId: undefined, hashKey: '$remote_addr' });
        editDirectives.value = [];
        hcType.value = 'none';
        hcIntervalSec.value = 5;
        hcPath.value = '/healthz';
        hcExpectedStatus.value = '200';
      }
      resetDirty();
    }
  },
});

async function onSubmit(values: Record<string, any>) {
  // 节点列表本地校验
  if (editNodes.value.length === 0) {
    message.warning('至少配置一个后端服务器节点');
    return;
  }
  for (const [index, node] of editNodes.value.entries()) {
    if (!node.host.trim()) {
      message.warning(`请填写第 ${index + 1} 个节点的地址`);
      return;
    }
    if (!node.port || node.port < 1 || node.port > 65_535) {
      message.warning(`第 ${index + 1} 个节点的端口必须是 1-65535`);
      return;
    }
  }
  if (editNodes.value.every((item) => item.backup)) {
    message.warning('不允许全部节点都标记为 backup，至少保留一个主节点');
    return;
  }
  if (hcType.value === 'http') {
    if (!hcPath.value.trim().startsWith('/')) {
      message.warning('HTTP 探测路径必须以 / 开头');
      return;
    }
    if (!hcExpectedStatus.value.trim()) {
      message.warning('HTTP 主动探测需配置期望状态码');
      return;
    }
  }

  const dirErr = validateDirectives();
  if (dirErr) {
    message.warning(dirErr);
    return;
  }

  const hashKey = String(values.hashKey ?? '').trim();
  if (
    (values.lbPolicy === 'hash' || values.lbPolicy === 'consistent_hash') &&
    !/^\$[A-Za-z_][A-Za-z0-9_]*(\$[A-Za-z_][A-Za-z0-9_]*)*$/.test(hashKey)
  ) {
    message.warning('Hash 键需填写 NGINX 变量，例如 $remote_addr');
    return;
  }

  const nodes: UpstreamNode[] = editNodes.value.map((item) => ({
    backup: item.backup,
    failTimeoutSec: item.failTimeoutSec,
    host: item.host.trim(),
    maxFails: item.maxFails,
    port: Number(item.port),
    slowStartSec: item.slowStartSec,
    weight: item.weight,
  }));
  const payload: UpstreamGroupPayload = {
    centerId: Number(values.centerId),
    description: String(values.description ?? ''),
    healthCheck: {
      expectedStatus:
        hcType.value === 'http'
          ? hcExpectedStatus.value
              .split(/[,，\s]+/)
              .map(Number)
              .filter(Number.isInteger)
          : [],
      intervalSec: hcIntervalSec.value,
      path: hcType.value === 'http' ? hcPath.value.trim() : '',
      type: hcType.value,
    },
    hashKey,
    directives: editDirectives.value.map((item) => ({
      name: item.name.trim(),
      value: item.value.trim(),
    })),
    lbPolicy: values.lbPolicy,
    name: String(values.name),
    nodes,
    tags: String(values.tagsText ?? '')
      .split(/[,，\s]+/)
      .map((item) => item.trim())
      .filter(Boolean),
  };

  modalApi.lock();
  try {
    if (isEdit.value) {
      await updateUpstreamGroupApi(editId.value, payload);
      message.success('上游组已更新');
    } else {
      await createUpstreamGroupApi(payload);
      message.success('上游组创建成功');
    }
    resetDirty();
    modalApi.close();
    emit('success');
  } catch {
    // 错误提示由请求拦截器统一处理
  } finally {
    modalApi.unlock();
  }
}
</script>

<template>
  <Modal class="w-[680px]">
    <div class="pr-2">
      <Form />
    </div>

    <!-- 后端节点列表 -->
    <div class="mt-4 px-1">
      <div class="mb-2 flex items-center justify-between">
        <span class="text-sm font-medium">后端服务器节点</span>
        <Button size="small" type="dashed" @click="addNode">+ 添加节点</Button>
      </div>
      <div class="max-h-64 space-y-2 overflow-y-auto pr-1">
        <div
          v-for="(node, index) in editNodes"
          :key="index"
          class="flex flex-wrap items-center gap-2 rounded border border-gray-200 bg-gray-50 p-2 dark:border-gray-700 dark:bg-gray-800"
        >
          <Tag class="shrink-0" color="blue">#{{ index + 1 }}</Tag>
          <Input
            v-model:value="node.host"
            class="!w-40"
            placeholder="地址，例如 10.60.1.11"
            size="small"
          />
          <InputNumber
            v-model:value="node.port"
            :max="65535"
            :min="1"
            class="!w-24"
            placeholder="端口"
            :precision="0"
            size="small"
          />
          <div class="flex items-center gap-1 text-xs text-gray-500">
            <span>权重</span>
            <InputNumber
              v-model:value="node.weight"
              :max="100"
              :min="1"
              class="!w-16"
              :precision="0"
              size="small"
            />
          </div>
          <div class="flex items-center gap-1 text-xs text-gray-500">
            <span>max_fails</span>
            <InputNumber
              v-model:value="node.maxFails"
              :max="100"
              :min="0"
              class="!w-16"
              :precision="0"
              size="small"
            />
          </div>
          <div class="flex items-center gap-1 text-xs text-gray-500">
            <span>fail_timeout</span>
            <InputNumber
              v-model:value="node.failTimeoutSec"
              :max="300"
              :min="1"
              class="!w-16"
              :precision="0"
              size="small"
            />
          </div>
          <Tooltip title="慢启动（秒），备节点常用">
            <div class="flex items-center gap-1 text-xs text-gray-500">
              <span>慢启动</span>
              <InputNumber
                v-model:value="node.slowStartSec"
                :max="600"
                :min="0"
                class="!w-16"
                :precision="0"
                size="small"
              />
            </div>
          </Tooltip>
          <Tooltip title="backup：仅当其他节点全部不可用时启用">
            <div class="flex items-center gap-1">
              <span class="text-xs text-gray-500">备用</span>
              <Switch
                v-model:checked="node.backup"
                checked-children="备"
                size="small"
                un-checked-children="主"
              />
            </div>
          </Tooltip>
          <Button danger size="small" type="link" @click="removeNode(index)">
            删除
          </Button>
        </div>
        <div
          v-if="editNodes.length === 0"
          class="py-4 text-center text-xs text-gray-400"
        >
          暂无节点，点击「+ 添加节点」开始配置
        </div>
      </div>
    </div>

    <!-- 健康检查 -->
    <div class="mt-4 px-1">
      <span class="text-sm font-medium">健康检查</span>
      <div class="mt-2 grid grid-cols-2 gap-3">
        <div>
          <div class="mb-1 text-xs text-gray-500">检查方式</div>
          <select
            v-model="hcType"
            class="h-8 w-full rounded border border-gray-300 bg-transparent px-2 text-sm dark:border-gray-600"
          >
            <option value="none">被动检查（nginx 原生 max_fails）</option>
            <option value="tcp">主动 TCP 探测</option>
            <option value="http">主动 HTTP 探测</option>
          </select>
        </div>
        <div>
          <div class="mb-1 text-xs text-gray-500">探测间隔（秒）</div>
          <InputNumber
            v-model:value="hcIntervalSec"
            :disabled="hcType === 'none'"
            :max="300"
            :min="1"
            class="w-full"
            :precision="0"
          />
        </div>
        <template v-if="hcType === 'http'">
          <div>
            <div class="mb-1 text-xs text-gray-500">探测路径</div>
            <Input v-model:value="hcPath" placeholder="/healthz" />
          </div>
          <div>
            <div class="mb-1 text-xs text-gray-500">期望状态码（逗号分隔）</div>
            <Input v-model:value="hcExpectedStatus" placeholder="200,204" />
          </div>
        </template>
      </div>
    </div>

    <!-- Upstream 块指令（upstream {} 上下文） -->
    <div class="mt-4 px-1">
      <div class="mb-2 flex items-center justify-between">
        <Tooltip
          title="upstream {} 块内指令（如 keepalive、keepalive_timeout、hash），仅对当前上游服务器组生效。server 节点已由上方表单托管，此处无需重复配置。"
        >
          <span class="text-sm font-medium">Upstream 块指令</span>
        </Tooltip>
        <Button size="small" type="dashed" @click="addDirective">+ 添加指令</Button>
      </div>
      <div class="max-h-60 space-y-2 overflow-y-auto pr-1">
        <div
          v-for="(directive, index) in editDirectives"
          :key="index"
          class="flex flex-wrap items-center gap-2 rounded border border-gray-200 bg-gray-50 p-2 dark:border-gray-700 dark:bg-gray-800"
        >
          <Tag class="shrink-0" color="blue">#{{ index + 1 }}</Tag>
          <AutoComplete
            v-model:value="directive.name"
            :filter-option="filterDirectiveOption"
            :options="UPSTREAM_DIRECTIVE_OPTIONS"
            class="!w-56"
            placeholder="指令名，例如 keepalive"
            size="small"
          />
          <Input
            v-model:value="directive.value"
            class="min-w-[140px] flex-1"
            placeholder="指令参数值，例如 32"
            size="small"
          />
          <Button danger size="small" type="link" @click="removeDirective(index)">
            删除
          </Button>
        </div>
        <div
          v-if="editDirectives.length === 0"
          class="py-3 text-center text-xs text-gray-400"
        >
          暂无 Upstream 块指令，可点击「+ 添加指令」配置 keepalive、hash 等
        </div>
      </div>
    </div>
  </Modal>
</template>
