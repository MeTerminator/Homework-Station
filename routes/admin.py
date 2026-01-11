from flask import Blueprint, render_template, redirect, url_for, flash, send_from_directory, current_app
from flask_login import login_required, current_user
import os
from models.user import User
from models.homework import Homework
from utils.helpers import get_homework_name
from extensions import db

admin_bp = Blueprint("admin", __name__)


@admin_bp.route("/admin/dashboard")
@login_required
def dashboard():
    if not current_user.is_admin:
        flash("无权限", "danger")
        return redirect(url_for("student.dashboard"))

    users = User.query.all()
    stats = {}
    total_files = 0
    total_size = 0  # 单位：字节

    # 计算作业数据
    for u in users:
        if not u.is_admin:
            hw_list: list[Homework] = Homework.query.filter_by(
                user_id=u.id).all()
            total_files += len(hw_list)  # 累加文件总数
            for h in hw_list:
                h.name = get_homework_name(h.homework_number)
            stats[u.id] = {
                "username": u.username,
                "student_id": u.student_id,
                "email": u.email,
                "homeworks": hw_list,
            }

    # 计算目录总大小
    upload_dir = current_app.config["UPLOAD_FOLDER"]
    if os.path.exists(upload_dir):
        for sid in os.listdir(upload_dir):
            for f in os.listdir(os.path.join(upload_dir, sid)):
                fp = os.path.join(upload_dir, sid, f)
                if os.path.isfile(fp):
                    total_size += os.path.getsize(fp)

    # 将字节转换为易读的格式 (MB)
    size_in_mb = round(total_size / (1024 * 1024), 2)

    return render_template(
        "admin_dashboard.html",
        user_stats=stats,
        total_files=total_files,
        total_size=size_in_mb
    )


@admin_bp.route("/admin/download/<int:homework_id>")
@login_required
def download(homework_id):
    hw = Homework.query.get_or_404(homework_id)
    if not current_user.is_admin:
        flash("无权限", "danger")
        return redirect(url_for("student.dashboard"))

    return send_from_directory(
        current_app.config["UPLOAD_FOLDER"],
        hw.file_path,
        as_attachment=True,
        download_name=hw.file_name,
    )


@admin_bp.route("/admin/delete_homework/<int:homework_id>")
@login_required
def delete(homework_id):
    if not current_user.is_admin:
        flash("无权限", "danger")
        return redirect(url_for("student.dashboard"))

    hw = Homework.query.get_or_404(homework_id)

    path = os.path.join(current_app.config["UPLOAD_FOLDER"], hw.file_path)
    if os.path.exists(path):
        os.remove(path)

    db.session.delete(hw)
    db.session.commit()

    flash("已删除", "success")
    return redirect(url_for("admin.dashboard"))
