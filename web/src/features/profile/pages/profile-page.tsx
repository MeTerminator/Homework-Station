import { PageHeading } from "@/components/page-heading";
import { BasicProfileCard } from "../components/basic-profile-card";
import { PasswordCard } from "../components/password-card";
import { QQBindingCard } from "../components/qq-binding-card";
import { RealNameBindingCard } from "../components/real-name-binding-card";
export function Profile() {
  return (
    <>
      <PageHeading
        title="个人信息"
        description="完善个人资料，让每一份作业都有清晰的归属。"
      />
      <div className="stack">
        <div className="grid-two">
          <BasicProfileCard />
          <PasswordCard />
        </div>
        <div className="grid-two">
          <RealNameBindingCard />
          <QQBindingCard />
        </div>
      </div>
    </>
  );
}
