// รองรับ FR-BI-01, NFR-PERF-02: รูปแบบข้อมูลจำนวนผู้ใช้บริการตามสัญญา API ใน specs/007-occupancy-dashboard/plan.md §4.1
export type OccupancySnapshot = {
  /** จำนวนผู้ใช้บริการภายในยิมตอนนี้ (DOM-CAP-01) */
  count: number;
  /** เวลาที่คำนวณจำนวนนี้ (RFC3339) */
  as_of: string;
};
