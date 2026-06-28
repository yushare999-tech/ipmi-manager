/**
 * IPMI Manager - Preload 스크립트
 * 작성일: 2026-06-25
 * 변경이력:
 *   - 2026-06-25: 최초 작성 (Java 탐지/관리 API 추가)
 *   - 2026-06-25: IPMI 자동 로그인 API 추가
 */

const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('ipmiAPI', {
  // ── 설정 관리 ──────────────────────────────────
  saveConfig: (config)          => ipcRenderer.invoke('config:save', config),
  loadConfig: ()                => ipcRenderer.invoke('config:load'),

  // ── KVM 접속 ────────────────────────────────────
  openHtml5Kvm:           (device)                       => ipcRenderer.invoke('kvm:open-html5', device),
  openHtml5KvmAutoLogin:  (device)                       => ipcRenderer.invoke('kvm:open-html5-autologin', device),
  launchExternal:         (device, method, javawsPath)   => ipcRenderer.invoke('kvm:launch-external', { device, method, javawsPath }),

  // ── IPMI 페이지 ──────────────────────────────────
  openIpmiAutoLogin:      (device)                       => ipcRenderer.invoke('ipmi:open-autologin', device),

  // ── Java 관리 ──────────────────────────────────
  detectJava:          ()          => ipcRenderer.invoke('java:detect'),
  addJavaException:    (siteUrl)   => ipcRenderer.invoke('java:add-exception', { siteUrl }),
  applyLegacyJava:     ()          => ipcRenderer.invoke('java:apply-legacy-config'),
  patchJavaSecurity:   (javawsPath) => ipcRenderer.invoke('java:patch-security', { javawsPath }),
  getJavaDownloadInfo: ()          => ipcRenderer.invoke('java:get-download-links'),

  // ── 시스템/유틸 ────────────────────────────────
  openUrl:    (url)     => ipcRenderer.invoke('shell:open-url', url),
  openFile:   (options) => ipcRenderer.invoke('dialog:open-file', options),
});
