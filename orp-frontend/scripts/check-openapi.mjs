import { openApiSpec } from '../apps/web-antd/src/api/orp/openapi-spec.ts';
import { readFile } from 'node:fs/promises';
import { resolve } from 'node:path';

const errors = [];
let operationCount = 0;
let exampleCount = 0;

function fail(path, message) {
  errors.push(`${path}: ${message}`);
}

function resolveLocalRef(ref) {
  if (!ref.startsWith('#/')) return undefined;
  return ref
    .slice(2)
    .split('/')
    .map((part) => part.replaceAll('~1', '/').replaceAll('~0', '~'))
    .reduce((value, part) => value?.[part], openApiSpec);
}

function validateExample(value, schema, location, seenRefs = new Set()) {
  if (!schema || typeof schema !== 'object') return;
  if (schema.$ref) {
    if (seenRefs.has(schema.$ref)) return;
    const target = resolveLocalRef(schema.$ref);
    if (!target) {
      fail(location, `无法解析引用 ${schema.$ref}`);
      return;
    }
    const nextRefs = new Set(seenRefs).add(schema.$ref);
    validateExample(value, target, location, nextRefs);
    return;
  }
  if (schema.nullable && value === null) return;
  if (schema.enum && !schema.enum.some((entry) => Object.is(entry, value))) {
    fail(location, `示例值 ${JSON.stringify(value)} 不属于 enum`);
    return;
  }

  const actualType = Array.isArray(value) ? 'array' : value === null ? 'null' : typeof value;
  const validType = {
    array: actualType === 'array',
    boolean: actualType === 'boolean',
    integer: Number.isInteger(value),
    number: typeof value === 'number' && Number.isFinite(value),
    object: actualType === 'object' && value !== null && !Array.isArray(value),
    string: actualType === 'string',
  }[schema.type];
  if (schema.type && !validType) {
    fail(location, `schema 声明为 ${schema.type}，示例实际为 ${actualType}`);
    return;
  }
  if (schema.type === 'object' && value && typeof value === 'object') {
    for (const field of schema.required ?? []) {
      if (!(field in value)) fail(location, `示例缺少必填字段 ${field}`);
    }
    for (const [field, fieldSchema] of Object.entries(schema.properties ?? {})) {
      if (field in value) validateExample(value[field], fieldSchema, `${location}.${field}`, seenRefs);
    }
  }
  if (schema.type === 'array' && Array.isArray(value) && schema.items) {
    value.forEach((entry, index) =>
      validateExample(entry, schema.items, `${location}[${index}]`, seenRefs),
    );
  }
  if (typeof value === 'number') {
    if (schema.minimum !== undefined && value < schema.minimum) {
      fail(location, `示例值小于 minimum ${schema.minimum}`);
    }
    if (schema.maximum !== undefined && value > schema.maximum) {
      fail(location, `示例值大于 maximum ${schema.maximum}`);
    }
  }
  if (typeof value === 'string') {
    if (schema.minLength !== undefined && value.length < schema.minLength) {
      fail(location, `示例长度小于 minLength ${schema.minLength}`);
    }
    if (schema.maxLength !== undefined && value.length > schema.maxLength) {
      fail(location, `示例长度大于 maxLength ${schema.maxLength}`);
    }
    if (schema.pattern && !new RegExp(schema.pattern).test(value)) {
      fail(location, `示例不符合 pattern ${schema.pattern}`);
    }
  }
  for (const branch of schema.allOf ?? []) {
    validateExample(value, branch, location, seenRefs);
  }
  for (const union of ['anyOf', 'oneOf']) {
    if (schema[union]) {
      let matchingBranches = 0;
      for (const branch of schema[union]) {
        const previousLength = errors.length;
        validateExample(value, branch, location, seenRefs);
        if (errors.length === previousLength) matchingBranches++;
        errors.splice(previousLength);
      }
      if (union === 'anyOf' && matchingBranches === 0) {
        fail(location, `示例不符合 ${union} 中的任何一个 schema`);
      }
      if (union === 'oneOf' && matchingBranches !== 1) {
        fail(location, `示例不符合 ${union} 中且仅中一个 schema`);
      }
    }
  }
}

function walkRefs(value, location, seen = new Set()) {
  if (!value || typeof value !== 'object' || seen.has(value)) return;
  seen.add(value);
  if (typeof value.$ref === 'string' && value.$ref.startsWith('#/')) {
    if (!resolveLocalRef(value.$ref)) fail(location, `无法解析引用 ${value.$ref}`);
  }
  for (const [key, child] of Object.entries(value)) {
    walkRefs(child, `${location}.${key}`, seen);
  }
}

walkRefs(openApiSpec, 'openapi');
const swaggerPath = resolve('apps/web-antd/public/swagger/openapi.json');
const swaggerSpec = JSON.parse(await readFile(swaggerPath, 'utf8'));
if (JSON.stringify(swaggerSpec) !== JSON.stringify(openApiSpec)) {
  fail('public/swagger/openapi.json', '静态 Swagger 文档与前端 OpenAPI 规范不一致；请运行 pnpm generate:openapi');
}
const methods = new Set(['delete', 'get', 'patch', 'post', 'put']);
const operationIds = new Set();
for (const [path, pathItem] of Object.entries(openApiSpec.paths ?? {})) {
  if (!path.startsWith('/')) fail(path, '路径必须以 / 开头');
  for (const [method, operation] of Object.entries(pathItem)) {
    if (!methods.has(method)) continue;
    operationCount++;
    const location = `${method.toUpperCase()} ${path}`;
    if (!operation.operationId) fail(location, '缺少 operationId');
    else if (operationIds.has(operation.operationId)) {
      fail(location, `operationId 重复：${operation.operationId}`);
    } else operationIds.add(operation.operationId);

    const pathParams = new Set([...path.matchAll(/\{([^}]+)\}/g)].map((match) => match[1]));
    const declaredPathParams = new Set(
      (operation.parameters ?? [])
        .filter((parameter) => parameter.in === 'path')
        .map((parameter) => parameter.name),
    );
    for (const parameter of pathParams) {
      if (!declaredPathParams.has(parameter)) fail(location, `缺少路径参数定义 ${parameter}`);
    }
    for (const parameter of declaredPathParams) {
      if (!pathParams.has(parameter)) fail(location, `路径参数 ${parameter} 未出现在 URL 中`);
    }
    for (const parameter of operation.parameters ?? []) {
      if (!parameter.schema && !parameter.content) {
        fail(location, `参数 ${parameter.name} 缺少 schema/content`);
      }
      if (Object.hasOwn(parameter.schema ?? {}, 'example')) {
        exampleCount++;
        validateExample(parameter.schema.example, parameter.schema, `${location} parameter ${parameter.name}`);
      }
    }
    for (const [status, response] of Object.entries(operation.responses ?? {})) {
      if (!response.description) fail(location, `响应 ${status} 缺少 description`);
      for (const [mediaType, media] of Object.entries(response.content ?? {})) {
        if (!media.schema) fail(location, `响应 ${status} ${mediaType} 缺少 schema`);
        if (Object.hasOwn(media, 'example')) {
          exampleCount++;
          validateExample(media.example, media.schema, `${location} response ${status} ${mediaType}`);
        }
        for (const [name, example] of Object.entries(media.examples ?? {})) {
          if (Object.hasOwn(example, 'value')) {
            exampleCount++;
            validateExample(example.value, media.schema, `${location} response ${status} example ${name}`);
          }
        }
      }
    }
    const requestContent = operation.requestBody?.content;
    if (requestContent) {
      for (const [mediaType, media] of Object.entries(requestContent)) {
        if (!media.schema) fail(location, `请求 ${mediaType} 缺少 schema`);
        if (Object.hasOwn(media, 'example')) {
          exampleCount++;
          validateExample(media.example, media.schema, `${location} request ${mediaType}`);
        }
      }
    }
  }
}

if (errors.length) {
  console.error(`OpenAPI 检查失败（${errors.length} 项）：`);
  for (const error of errors) console.error(`- ${error}`);
  process.exitCode = 1;
} else {
  console.log(`OpenAPI 检查通过：${Object.keys(openApiSpec.paths).length} 条路径，${operationCount} 个操作，${exampleCount} 个示例已检查。`);
}
