import { writeFile } from 'node:fs/promises';
import { resolve } from 'node:path';

import { openApiSpec } from '../apps/web-antd/src/api/orp/openapi-spec.ts';

const destination = resolve('apps/web-antd/public/swagger/openapi.json');
await writeFile(destination, `${JSON.stringify(openApiSpec, null, 2)}\n`);
console.log(`Generated ${destination}`);
