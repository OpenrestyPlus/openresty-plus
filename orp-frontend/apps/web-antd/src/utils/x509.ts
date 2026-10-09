/**
 * 轻量 X.509 证书解析器（无三方依赖）：
 * 解析 PEM 公钥证书，提取有效期（notBefore/notAfter）与关联域名（CN + SAN dNSName）。
 * 仅覆盖 DER TLV 遍历所需的 ASN.1 子集，不做签名校验。
 */

export interface ParsedCertificate {
  /** 主题备用名称 dNSName 列表 */
  sanDns: string[];
  /** 主题通用名称 */
  subjectCn: string;
  /** 有效期起始（YYYY-MM-DD HH:mm:ss） */
  notAfter: string;
  notBefore: string;
}

const TAG_INT = 0x02;
const TAG_SEQ = 0x30;
const TAG_SET = 0x31;
const TAG_OID = 0x06;
const TAG_UTCTIME = 0x17;
const TAG_GENTIME = 0x18;
/** CN OID 2.5.29.17 → 55 04 03；CN 是 55 04 03 */
const CN_OID_HEX = '550403';
/** subjectAltName OID 2.5.29.17 */
const SAN_OID = [0x55, 0x1d, 0x11];

/** PEM → DER（Base64 解码） */
function pemToDer(pem: string): Uint8Array | null {
  const match = pem.match(
    /-----BEGIN CERTIFICATE-----([\s\S]*?)-----END CERTIFICATE-----/,
  );
  if (!match) return null;
  const b64 = match[1]?.replaceAll(/[^A-Za-z0-9+/=]/g, '') ?? '';
  try {
    const bin = atob(b64);
    const der = new Uint8Array(bin.length);
    for (let i = 0; i < bin.length; i++) der[i] = bin.charCodeAt(i);
    return der;
  } catch {
    return null;
  }
}

interface Tlv {
  tag: number;
  value: Uint8Array;
  /** value 在原 DER 中的偏移 */
  start: number;
}

/** 读取一个 TLV，返回 [tlv, 下一字节偏移]；失败返回 null */
function readTlv(der: Uint8Array, offset: number): [Tlv, number] | null {
  if (offset + 2 > der.length) return null;
  const tag = der[offset] as number;
  let len = der[offset + 1] as number;
  let headerLen = 2;
  if (len & 0x80) {
    const numBytes = len & 0x07;
    if (numBytes === 0 || offset + 2 + numBytes > der.length) return null;
    len = 0;
    for (let i = 0; i < numBytes; i++) {
      len = len * 256 + (der[offset + 2 + i] as number);
    }
    headerLen = 2 + numBytes;
  }
  const start = offset + headerLen;
  if (start + len > der.length) return null;
  return [
    { tag, value: der.subarray(start, start + len), start },
    start + len,
  ];
}

/** 遍历 SET/SEQ 的直接子元素 */
function* children(der: Uint8Array, tlv: Tlv) {
  let offset = tlv.start;
  const end = tlv.start + tlv.value.length;
  while (offset < end) {
    const next = readTlv(der, offset);
    if (!next) return;
    yield next[0];
    offset = next[1];
  }
}

/** 解析 ASN.1 时间（UTCTime YYMMDDHHMMSSZ / GeneralizedTime YYYYMMDDHHMMSSZ） */
function parseAsn1Time(bytes: Uint8Array): string | null {
  const s = new TextDecoder().decode(bytes);
  const m = s.match(/^(\d{2}|\d{4})(\d{2})(\d{2})(\d{2})(\d{2})(\d{2})?Z?$/);
  if (!m) return null;
  let year = Number(m[1]);
  if (m[1]?.length === 2) {
    year = year >= 50 ? 1900 + year : 2000 + year;
  }
  const pad = (n: number | undefined) => String(n ?? 0).padStart(2, '0');
  return `${year}-${pad(Number(m[2]))}-${pad(Number(m[3]))} ${pad(Number(m[4]))}:${pad(Number(m[5]))}:${pad(Number(m[6]))}`;
}

/** 解析 Name（RDNSequence）取 CN */
function parseName(der: Uint8Array, nameSeq: Tlv): null | string {
  let cn: null | string = null;
  for (const rdn of children(der, nameSeq)) {
    if (rdn.tag !== TAG_SET) continue;
    for (const atv of children(der, rdn)) {
      if (atv.tag !== TAG_SEQ) continue;
      const [oid, value] = [...children(der, atv)];
      if (!oid || !value || oid.tag !== TAG_OID) continue;
      const hex = Array.from(oid.value)
        .map((b) => b.toString(16).padStart(2, '0'))
        .join('');
      if (hex === CN_OID_HEX) {
        cn = new TextDecoder().decode(value.value);
      }
    }
  }
  return cn;
}

/** 解析 SubjectAltName（OCTET STRING 内容 = GeneralNames SEQ）→ dNSName 列表 */
function parseSan(der: Uint8Array, octets: Tlv): string[] {
  const dns: string[] = [];
  const generalNames = readTlv(der, octets.start);
  if (!generalNames || generalNames[0].tag !== TAG_SEQ) return dns;
  for (const gname of children(der, generalNames[0])) {
    // [2] IMPLICIT dNSName
    if (gname.tag === 0x82) {
      dns.push(new TextDecoder().decode(gname.value));
    }
  }
  return dns;
}

/** 解析证书：返回 notBefore/notAfter/CN/SAN；失败返回 null */
export function parseCertificate(pem: string): ParsedCertificate | null {
  const der = pemToDer(pem);
  if (!der) return null;
  const root = readTlv(der, 0);
  if (!root || root[0].tag !== TAG_SEQ) return null;
  const [tbs] = [...children(der, root[0])];
  if (!tbs || tbs.tag !== TAG_SEQ) return null;

  const result: ParsedCertificate = {
    notAfter: '',
    notBefore: '',
    sanDns: [],
    subjectCn: '',
  };

  for (const field of children(der, tbs)) {
    if (field.tag === TAG_INT || field.tag === TAG_SET) continue;

    // validity / issuer / subject / SPKI 都是 SEQ
    if (field.tag === TAG_SEQ) {
      const kids = [...children(der, field)];
      const times = kids.filter(
        (k) => k.tag === TAG_UTCTIME || k.tag === TAG_GENTIME,
      );
      if (times.length === 2) {
        // validity SEQ：两个时间子元素
        result.notBefore = parseAsn1Time(times[0]!.value) ?? '';
        result.notAfter = parseAsn1Time(times[1]!.value) ?? '';
      } else {
        // issuer / subject Name：后者覆盖前者（subject 在后）
        const cn = parseName(der, field);
        if (cn) result.subjectCn = cn;
      }
      continue;
    }

    // extensions [3] EXPLICIT
    if (field.tag === 0xa3) {
      for (const exts of children(der, field)) {
        if (exts.tag !== TAG_SEQ) continue;
        for (const ext of children(der, exts)) {
          if (ext.tag !== TAG_SEQ) continue;
          const parts = [...children(der, ext)];
          const oid = parts[0];
          if (!oid || oid.tag !== TAG_OID) continue;
          const oidBytes = Array.from(oid.value);
          if (
            oidBytes.length === 3 &&
            oidBytes[0] === SAN_OID[0] &&
            oidBytes[1] === SAN_OID[1] &&
            oidBytes[2] === SAN_OID[2]
          ) {
            const octets = parts.find((p) => p.tag === 0x04);
            if (octets) {
              result.sanDns = parseSan(der, octets);
            }
          }
        }
      }
    }
  }

  if (!result.notBefore || !result.notAfter) return null;
  return result;
}
