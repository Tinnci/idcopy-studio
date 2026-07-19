import { IDCopyService } from "../bindings/github.com/Tinnci/idcopy-studio/index.js";
import type {
  Catalog,
  DocumentSpec,
  LayoutPlan,
  LayoutRequest,
  PaperSpec,
} from "../bindings/github.com/Tinnci/idcopy-studio/idcopy/models.js";

type Side = "front" | "back";

interface WatermarkConfig {
  enabled: boolean;
  text: string;
  mode: "repeat" | "center";
  color: string;
  opacity: number;
  sizePercent: number;
  angle: number;
}

const byID = <T extends HTMLElement>(id: string): T => {
  const element = document.getElementById(id);
  if (!element) throw new Error(`Missing element #${id}`);
  return element as T;
};

const documentSelect = byID<HTMLSelectElement>("document-select");
const paperSelect = byID<HTMLSelectElement>("paper-select");
const sidesSelect = byID<HTMLSelectElement>("sides-select");
const orientationSelect = byID<HTMLSelectElement>("orientation-select");
const arrangementSelect = byID<HTMLSelectElement>("arrangement-select");
const fitSelect = byID<HTMLSelectElement>("fit-select");
const copiesInput = byID<HTMLInputElement>("copies-input");
const scaleInput = byID<HTMLInputElement>("scale-input");
const marginInput = byID<HTMLInputElement>("margin-input");
const gapInput = byID<HTMLInputElement>("gap-input");
const cutmarksInput = byID<HTMLInputElement>("cutmarks-input");
const filenameInput = byID<HTMLInputElement>("filename-input");
const qualitySelect = byID<HTMLSelectElement>("quality-select");
const encodingSelect = byID<HTMLSelectElement>("encoding-select");
const watermarkInput = byID<HTMLInputElement>("watermark-input");
const watermarkSettings = byID<HTMLDivElement>("watermark-settings");
const watermarkText = byID<HTMLInputElement>("watermark-text");
const watermarkMode = byID<HTMLSelectElement>("watermark-mode");
const watermarkColor = byID<HTMLSelectElement>("watermark-color");
const watermarkOpacity = byID<HTMLInputElement>("watermark-opacity");
const watermarkSize = byID<HTMLInputElement>("watermark-size");
const watermarkAngle = byID<HTMLInputElement>("watermark-angle");
const watermarkOpacityOutput = byID<HTMLOutputElement>("watermark-opacity-output");
const watermarkSizeOutput = byID<HTMLOutputElement>("watermark-size-output");
const pdfTitle = byID<HTMLInputElement>("pdf-title");
const pdfAuthor = byID<HTMLInputElement>("pdf-author");
const pdfSubject = byID<HTMLInputElement>("pdf-subject");
const pdfKeywords = byID<HTMLInputElement>("pdf-keywords");
const documentWidthInput = byID<HTMLInputElement>("document-width");
const documentHeightInput = byID<HTMLInputElement>("document-height");
const paperWidthInput = byID<HTMLInputElement>("paper-width");
const paperHeightInput = byID<HTMLInputElement>("paper-height");
const documentMeta = byID<HTMLParagraphElement>("document-meta");
const customDocument = byID<HTMLDivElement>("custom-document");
const customPaper = byID<HTMLDivElement>("custom-paper");
const paperPreview = byID<HTMLDivElement>("paper-preview");
const previewTitle = byID<HTMLHeadingElement>("preview-title");
const documentSizeStat = byID<HTMLSpanElement>("document-size-stat");
const pageCountStat = byID<HTMLSpanElement>("page-count-stat");
const orientationNote = byID<HTMLParagraphElement>("orientation-note");
const pageIndicator = byID<HTMLSpanElement>("page-indicator");
const previousPage = byID<HTMLButtonElement>("previous-page");
const nextPage = byID<HTMLButtonElement>("next-page");
const exportButton = byID<HTMLButtonElement>("export-button");
const downloadButton = byID<HTMLButtonElement>("download-button");
const statusText = byID<HTMLParagraphElement>("status-text");

const dataURLs: Record<Side, string> = { front: "", back: "" };
const rotations: Record<Side, number> = { front: 0, back: 0 };
let catalog: Catalog = { papers: [], documents: [] };
let plan: LayoutPlan | null = null;
let currentPage = 1;
let previewSequence = 0;
let previewTimer = 0;

function numberValue(input: HTMLInputElement, fallback: number): number {
  const value = Number(input.value);
  return Number.isFinite(value) ? value : fallback;
}

function buildRequest(): LayoutRequest {
  return {
    paperID: paperSelect.value,
    customPaper: {
      width: numberValue(paperWidthInput, 210),
      height: numberValue(paperHeightInput, 297),
    },
    documentID: documentSelect.value,
    customDocument: {
      width: numberValue(documentWidthInput, 85.6),
      height: numberValue(documentHeightInput, 54),
    },
    orientation: orientationSelect.value,
    arrangement: arrangementSelect.value,
    sides: sidesSelect.value,
    copies: numberValue(copiesInput, 1),
    scalePercent: numberValue(scaleInput, 100),
    marginMM: numberValue(marginInput, 12),
    gapMM: numberValue(gapInput, 8),
    fit: fitSelect.value,
    cutMarks: cutmarksInput.checked,
    front: { dataURL: dataURLs.front, rotation: rotations.front },
    back: { dataURL: dataURLs.back, rotation: rotations.back },
    pdf: {
      title: pdfTitle.value.trim(),
      author: pdfAuthor.value.trim(),
      subject: pdfSubject.value.trim(),
      keywords: pdfKeywords.value.trim(),
    },
  };
}

function readWatermark(): WatermarkConfig {
  return {
    enabled: watermarkInput.checked,
    text: watermarkText.value.trim() || "仅供复印使用",
    mode: watermarkMode.value === "center" ? "center" : "repeat",
    color: watermarkColor.value,
    opacity: numberValue(watermarkOpacity, 28) / 100,
    sizePercent: numberValue(watermarkSize, 7),
    angle: numberValue(watermarkAngle, -28),
  };
}

function groupOptions<T extends { id: string; name: string }>(
  select: HTMLSelectElement,
  items: T[],
  groupOf: (item: T) => string,
): void {
  select.replaceChildren();
  const groups = new Map<string, T[]>();
  for (const item of items) {
    const group = groupOf(item);
    groups.set(group, [...(groups.get(group) ?? []), item]);
  }
  for (const [label, members] of groups) {
    const optgroup = document.createElement("optgroup");
    optgroup.label = label;
    for (const item of members) {
      const option = document.createElement("option");
      option.value = item.id;
      option.textContent = item.name;
      optgroup.append(option);
    }
    select.append(optgroup);
  }
  const custom = document.createElement("option");
  custom.value = "custom";
  custom.textContent = "自定义尺寸…";
  select.append(custom);
}

function selectedDocument(): DocumentSpec | undefined {
  return (catalog.documents ?? []).find((item) => item.id === documentSelect.value);
}

function selectedPaper(): PaperSpec | undefined {
  return (catalog.papers ?? []).find((item) => item.id === paperSelect.value);
}

function updateConditionalFields(): void {
  customDocument.classList.toggle("hidden", documentSelect.value !== "custom");
  customPaper.classList.toggle("hidden", paperSelect.value !== "custom");
  watermarkSettings.classList.toggle("hidden", !watermarkInput.checked);
  watermarkOpacityOutput.value = `${watermarkOpacity.value}%`;
  watermarkSizeOutput.value = `${watermarkSize.value}%`;
  const document = selectedDocument();
  documentMeta.textContent = document
    ? `${document.standard} · ${formatMM(document.width)} × ${formatMM(document.height)} mm${document.notes ? ` · ${document.notes}` : ""}`
    : "输入实物测量尺寸；旧版或非标准证件建议使用自定义尺寸。";

  const activeSides = sidesSelect.value;
  byID<HTMLDivElement>("front-slot").classList.toggle("inactive", activeSides === "back");
  byID<HTMLDivElement>("back-slot").classList.toggle("inactive", activeSides === "front");
}

function formatMM(value: number): string {
  return Number.isInteger(value) ? String(value) : value.toFixed(2).replace(/0+$/, "").replace(/\.$/, "");
}

function schedulePreview(): void {
  window.clearTimeout(previewTimer);
  previewTimer = window.setTimeout(() => void refreshPreview(), 120);
}

async function refreshPreview(): Promise<void> {
  updateConditionalFields();
  const sequence = ++previewSequence;
  try {
    const nextPlan = await IDCopyService.Preview(buildRequest());
    if (sequence !== previewSequence) return;
    plan = nextPlan;
    currentPage = Math.min(Math.max(currentPage, 1), nextPlan.pages);
    renderPreview();
    setStatus(requiredImagesReady() ? "可以生成 PDF" : "请选择所需的证件图像", "neutral");
  } catch (error) {
    if (sequence !== previewSequence) return;
    plan = null;
    renderPreviewError(errorMessage(error));
  }
}

function renderPreview(): void {
  if (!plan) return;
  const request = buildRequest();
  paperPreview.replaceChildren();
  paperPreview.style.aspectRatio = `${plan.paper.width} / ${plan.paper.height}`;
  const items = (plan.items ?? []).filter((item) => item.page === currentPage);

  for (const item of items) {
    const element = document.createElement("div");
    element.className = `placed-document ${item.side}`;
    element.style.left = `${(item.x / plan.paper.width) * 100}%`;
    element.style.top = `${(item.y / plan.paper.height) * 100}%`;
    element.style.width = `${(item.width / plan.paper.width) * 100}%`;
    element.style.height = `${(item.height / plan.paper.height) * 100}%`;

    const side = item.side as Side;
    if (dataURLs[side]) {
      const image = document.createElement("img");
      image.src = dataURLs[side];
      image.alt = side === "front" ? "正面" : "背面";
      image.style.objectFit = request.fit === "stretch" ? "fill" : request.fit;
      image.style.transform = `rotate(${rotations[side]}deg)`;
      element.append(image);
    } else {
      const placeholder = document.createElement("div");
      placeholder.className = "placed-placeholder";
      placeholder.textContent = side === "front" ? "正面" : "背面";
      element.append(placeholder);
    }

    const badge = document.createElement("span");
    badge.className = "side-badge";
    badge.textContent = `${side === "front" ? "正" : "背"} · ${item.copy}`;
    element.append(badge);
    appendWatermarkPreview(element);
    paperPreview.append(element);
  }

  previewTitle.textContent = `${selectedPaper()?.name ?? "自定义纸张"} · ${formatMM(plan.paper.width)} × ${formatMM(plan.paper.height)} mm`;
  documentSizeStat.textContent = `证件 ${formatMM(plan.document.width)} × ${formatMM(plan.document.height)} mm`;
  pageCountStat.textContent = `${plan.pages} 页 · ${plan.items?.length ?? 0} 面`;
  orientationNote.textContent = `${orientationSelect.value === "auto" ? "自动方向" : "指定方向"}：${plan.orientation === "landscape" ? "横向" : "纵向"}`;
  pageIndicator.textContent = `第 ${currentPage} / ${plan.pages} 页`;
  previousPage.disabled = currentPage <= 1;
  nextPage.disabled = currentPage >= plan.pages;
}

function appendWatermarkPreview(element: HTMLElement): void {
  const watermark = readWatermark();
  if (!watermark.enabled) return;
  const layer = document.createElement("div");
  layer.className = `watermark-preview ${watermark.mode}`;
  layer.style.setProperty("--watermark-color", watermark.color);
  layer.style.setProperty("--watermark-opacity", String(watermark.opacity));
  layer.style.setProperty("--watermark-angle", `${watermark.angle}deg`);
  layer.style.setProperty("--watermark-size", `${watermark.sizePercent * 0.24}cqw`);
  const repetitions = watermark.mode === "repeat" ? 12 : 1;
  for (let index = 0; index < repetitions; index++) {
    const label = document.createElement("span");
    label.textContent = watermark.text;
    layer.append(label);
  }
  element.append(layer);
}

function renderPreviewError(message: string): void {
  paperPreview.replaceChildren();
  paperPreview.style.aspectRatio = "210 / 297";
  const error = document.createElement("div");
  error.className = "preview-error";
  error.innerHTML = `<strong>无法排版</strong><span>${escapeHTML(message)}</span>`;
  paperPreview.append(error);
  pageCountStat.textContent = "—";
  pageIndicator.textContent = "第 0 / 0 页";
  previousPage.disabled = true;
  nextPage.disabled = true;
  setStatus(message, "error");
}

function requiredImagesReady(): boolean {
  if (sidesSelect.value === "front") return Boolean(dataURLs.front);
  if (sidesSelect.value === "back") return Boolean(dataURLs.back);
  return Boolean(dataURLs.front && dataURLs.back);
}

function setStatus(message: string, tone: "neutral" | "error" | "success"): void {
  statusText.textContent = message;
  statusText.dataset.tone = tone;
}

function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}

function escapeHTML(value: string): string {
  const container = document.createElement("div");
  container.textContent = value;
  return container.innerHTML;
}

async function loadImage(side: Side, file: File): Promise<void> {
  if (!file.type.match(/^image\/(jpeg|png)$/)) {
    setStatus("仅支持 JPEG 或 PNG 图像", "error");
    return;
  }
  if (file.size > 48 * 1024 * 1024) {
    setStatus("单张图像不能超过 48 MiB", "error");
    return;
  }
  const dataURL = await new Promise<string>((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(String(reader.result));
    reader.onerror = () => reject(reader.error ?? new Error("读取图像失败"));
    reader.readAsDataURL(file);
  });
  dataURLs[side] = dataURL;
  rotations[side] = 0;
  const thumb = byID<HTMLImageElement>(`${side}-thumb`);
  thumb.src = dataURL;
  thumb.closest(".image-slot")?.classList.add("has-image");
  setStatus(`${side === "front" ? "正面" : "背面"}已载入：${file.name}`, "success");
  schedulePreview();
}

function configureImageSlot(side: Side): void {
  const slot = byID<HTMLDivElement>(`${side}-slot`);
  const input = byID<HTMLInputElement>(`${side}-file`);
  const rotate = byID<HTMLButtonElement>(`${side}-rotate`);

  slot.addEventListener("click", (event) => {
    if (event.target !== rotate) input.click();
  });
  slot.addEventListener("keydown", (event) => {
    if (event.key === "Enter" || event.key === " ") input.click();
  });
  input.addEventListener("change", () => {
    const file = input.files?.[0];
    if (file) void loadImage(side, file);
  });
  for (const eventName of ["dragenter", "dragover"]) {
    slot.addEventListener(eventName, (event) => {
      event.preventDefault();
      slot.classList.add("dragging");
    });
  }
  for (const eventName of ["dragleave", "drop"]) {
    slot.addEventListener(eventName, (event) => {
      event.preventDefault();
      slot.classList.remove("dragging");
    });
  }
  slot.addEventListener("drop", (event) => {
    const file = event.dataTransfer?.files[0];
    if (file) void loadImage(side, file);
  });
  rotate.addEventListener("click", (event) => {
    event.stopPropagation();
    if (!dataURLs[side]) return;
    rotations[side] = (rotations[side] + 90) % 360;
    byID<HTMLImageElement>(`${side}-thumb`).style.transform = `rotate(${rotations[side]}deg)`;
    schedulePreview();
  });
}

function base64PDF(data: string): Blob {
  const chunks: ArrayBuffer[] = [];
  const chunkSize = 32_768;
  for (let offset = 0; offset < data.length; offset += chunkSize) {
    const decoded = atob(data.slice(offset, offset + chunkSize));
    const buffer = new ArrayBuffer(decoded.length);
    const bytes = new Uint8Array(buffer);
    for (let index = 0; index < decoded.length; index++) bytes[index] = decoded.charCodeAt(index);
    chunks.push(buffer);
  }
  return new Blob(chunks, { type: "application/pdf" });
}

function qualityDPI(): number {
  switch (qualitySelect.value) {
    case "compact": return 150;
    case "print": return 300;
    case "archive": return 600;
    default: return 0;
  }
}

function jpegQuality(): number {
  switch (qualitySelect.value) {
    case "compact": return 0.78;
    case "archive": return 0.95;
    default: return 0.9;
  }
}

async function loadImageElement(dataURL: string): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const image = new Image();
    image.onload = () => resolve(image);
    image.onerror = () => reject(new Error("无法解码待导出的图像"));
    image.src = dataURL;
  });
}

function drawFlattenedWatermark(context: CanvasRenderingContext2D, width: number, height: number, watermark: WatermarkConfig): void {
  if (!watermark.enabled) return;
  const fontSize = Math.max(12, Math.min(width, height) * watermark.sizePercent / 100);
  const diagonal = Math.hypot(width, height);
  context.save();
  context.translate(width / 2, height / 2);
  context.rotate(watermark.angle * Math.PI / 180);
  context.globalAlpha = watermark.opacity;
  context.fillStyle = watermark.color;
  context.font = `700 ${fontSize}px sans-serif`;
  context.textAlign = "center";
  context.textBaseline = "middle";
  if (watermark.mode === "center") {
    context.fillText(watermark.text, 0, 0);
  } else {
    const textWidth = Math.max(context.measureText(watermark.text).width, fontSize * 4);
    const stepX = textWidth + fontSize * 2.2;
    const stepY = fontSize * 3.2;
    let row = 0;
    for (let y = -diagonal; y <= diagonal; y += stepY) {
      const offset = row % 2 === 0 ? 0 : stepX / 2;
      for (let x = -diagonal; x <= diagonal; x += stepX) {
        context.fillText(watermark.text, x + offset, y);
      }
      row++;
    }
  }
  context.restore();
}

async function processImageForExport(
  dataURL: string,
  rotation: number,
  documentWidthMM: number,
  documentHeightMM: number,
  watermark: WatermarkConfig,
): Promise<string> {
  const dpi = qualityDPI();
  if (dpi === 0 && !watermark.enabled && rotation === 0) return dataURL;

  const image = await loadImageElement(dataURL);
  const rotated = rotation % 180 !== 0;
  const orientedWidth = rotated ? image.naturalHeight : image.naturalWidth;
  const orientedHeight = rotated ? image.naturalWidth : image.naturalHeight;
  let scale = 1;
  if (dpi > 0) {
    const maximumWidth = Math.max(1, Math.round(documentWidthMM / 25.4 * dpi));
    const maximumHeight = Math.max(1, Math.round(documentHeightMM / 25.4 * dpi));
    scale = Math.min(scale, maximumWidth / orientedWidth, maximumHeight / orientedHeight);
  }
  const maximumPixels = 50_000_000;
  if (orientedWidth * orientedHeight * scale * scale > maximumPixels) {
    scale = Math.sqrt(maximumPixels / (orientedWidth * orientedHeight));
  }
  const width = Math.max(1, Math.round(orientedWidth * scale));
  const height = Math.max(1, Math.round(orientedHeight * scale));
  const canvas = document.createElement("canvas");
  canvas.width = width;
  canvas.height = height;
  const context = canvas.getContext("2d", { alpha: false });
  if (!context) throw new Error("当前 WebView 不支持图像导出画布");
  context.fillStyle = "#ffffff";
  context.fillRect(0, 0, width, height);
  context.save();
  context.translate(width / 2, height / 2);
  context.rotate(rotation * Math.PI / 180);
  context.drawImage(
    image,
    -image.naturalWidth * scale / 2,
    -image.naturalHeight * scale / 2,
    image.naturalWidth * scale,
    image.naturalHeight * scale,
  );
  context.restore();
  drawFlattenedWatermark(context, width, height, watermark);
  if (encodingSelect.value === "png") return canvas.toDataURL("image/png");
  return canvas.toDataURL("image/jpeg", jpegQuality());
}

async function prepareExportRequest(): Promise<LayoutRequest> {
  const request = buildRequest();
  const exportPlan = plan ?? await IDCopyService.Preview(request);
  const watermark = readWatermark();
  const jobs: Promise<void>[] = [];
  if (request.sides !== "back") {
    jobs.push(processImageForExport(request.front.dataURL, request.front.rotation, exportPlan.document.width, exportPlan.document.height, watermark).then((dataURL) => {
      request.front = { dataURL, rotation: 0 };
    }));
  }
  if (request.sides !== "front") {
    jobs.push(processImageForExport(request.back.dataURL, request.back.rotation, exportPlan.document.width, exportPlan.document.height, watermark).then((dataURL) => {
      request.back = { dataURL, rotation: 0 };
    }));
  }
  await Promise.all(jobs);
  return request;
}

function setExportBusy(busy: boolean): void {
  exportButton.disabled = busy;
  downloadButton.disabled = busy;
  exportButton.classList.toggle("loading", busy);
}

function outputFilename(): string {
  const value = filenameInput.value.trim() || "id-copy.pdf";
  return value.toLowerCase().endsWith(".pdf") ? value : `${value}.pdf`;
}

async function savePDF(): Promise<void> {
  if (!requiredImagesReady()) {
    setStatus("请先载入所选输出面对应的图像", "error");
    return;
  }
  setExportBusy(true);
  setStatus(watermarkInput.checked ? "正在压平水印并生成 PDF…" : "正在生成 PDF…", "neutral");
  try {
    const request = await prepareExportRequest();
    const result = await IDCopyService.Save(request, outputFilename());
    if (result.canceled) {
      setStatus("已取消保存", "neutral");
      return;
    }
    setStatus(`已保存 ${result.pages} 页 · ${(result.bytes / 1024).toFixed(0)} KiB · ${result.path}`, "success");
  } catch (error) {
    setStatus(errorMessage(error), "error");
  } finally {
    setExportBusy(false);
  }
}

async function downloadPDF(): Promise<void> {
  if (!requiredImagesReady()) {
    setStatus("请先载入所选输出面对应的图像", "error");
    return;
  }
  setExportBusy(true);
  setStatus("正在生成浏览器下载副本…", "neutral");
  try {
    const request = await prepareExportRequest();
    const result = await IDCopyService.Generate(request);
    const url = URL.createObjectURL(base64PDF(result.dataBase64));
    const link = document.createElement("a");
    link.href = url;
    link.download = outputFilename() || result.filename;
    link.click();
    window.setTimeout(() => URL.revokeObjectURL(url), 30_000);
    setStatus(`已生成 ${result.pages} 页 · ${(result.bytes / 1024).toFixed(0)} KiB`, "success");
  } catch (error) {
    setStatus(errorMessage(error), "error");
  } finally {
    setExportBusy(false);
  }
}

async function initialise(): Promise<void> {
  try {
    catalog = await IDCopyService.Catalog();
    groupOptions(documentSelect, catalog.documents ?? [], (item) => item.region);
    groupOptions(paperSelect, catalog.papers ?? [], (item) => item.group);
    documentSelect.value = "cn-resident-id";
    paperSelect.value = "iso-a4";
    updateConditionalFields();
    await refreshPreview();
  } catch (error) {
    renderPreviewError(errorMessage(error));
  }
}

configureImageSlot("front");
configureImageSlot("back");

const reactiveControls: HTMLElement[] = [
  documentSelect, paperSelect, sidesSelect, orientationSelect, arrangementSelect, fitSelect,
  copiesInput, scaleInput, marginInput, gapInput, cutmarksInput,
  documentWidthInput, documentHeightInput, paperWidthInput, paperHeightInput,
  qualitySelect, encodingSelect, watermarkInput, watermarkText, watermarkMode, watermarkColor,
  watermarkOpacity, watermarkSize, watermarkAngle, pdfTitle, pdfAuthor, pdfSubject, pdfKeywords,
];
for (const control of reactiveControls) {
  control.addEventListener("input", schedulePreview);
  control.addEventListener("change", schedulePreview);
}

previousPage.addEventListener("click", () => {
  if (!plan || currentPage <= 1) return;
  currentPage--;
  renderPreview();
});
nextPage.addEventListener("click", () => {
  if (!plan || currentPage >= plan.pages) return;
  currentPage++;
  renderPreview();
});
exportButton.addEventListener("click", () => void savePDF());
downloadButton.addEventListener("click", () => void downloadPDF());

void initialise();
