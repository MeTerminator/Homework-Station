from flask import Blueprint, render_template, redirect, url_for, flash, current_app, send_from_directory
from flask_login import login_required, current_user
import os
from forms.homework import HomeworkForm
from models.homework import Homework
from utils.helpers import allowed_file, secure_filename, get_homework_name
from extensions import db

student_bp = Blueprint("student", __name__)


@student_bp.route("/student/dashboard")
@login_required
def dashboard():
    if current_user.is_admin:
        return redirect(url_for("admin.dashboard"))

    form = HomeworkForm()
    list_hw = Homework.query.filter_by(user_id=current_user.id).all()
    for h in list_hw:
        h.name = get_homework_name(h.homework_number)

    return render_template("student_dashboard.html", form=form, homeworks=list_hw)


@student_bp.route("/homework/submit", methods=["POST"])
@login_required
def submit():
    if current_user.is_admin:
        return redirect(url_for("admin.dashboard"))

    form = HomeworkForm()
    if form.validate_on_submit():
        file = form.file.data
        num = form.homework_number.data

        if not allowed_file(file.filename):
            flash("文件格式不支持，只允许 zip / pdf / md", "danger")
            return redirect(url_for("student.dashboard"))

        user_folder = os.path.join(
            current_app.config["UPLOAD_FOLDER"], current_user.student_id)
        os.makedirs(user_folder, exist_ok=True)

        safe_name = secure_filename(file.filename)
        save_name = f"{num}_{safe_name}"
        save_path = os.path.join(user_folder, save_name)

        exist = Homework.query.filter_by(
            user_id=current_user.id, homework_number=num).first()
        if exist:
            old_file = os.path.join(
                current_app.config["UPLOAD_FOLDER"], exist.file_path)
            if os.path.exists(old_file):
                os.remove(old_file)
            exist.file_name = file.filename
            exist.file_path = f"{current_user.student_id}/{save_name}"
        else:
            hw = Homework(
                user_id=current_user.id,
                homework_number=num,
                file_name=file.filename,
                file_path=f"{current_user.student_id}/{save_name}",
            )
            db.session.add(hw)

        file.save(save_path)
        db.session.commit()

        flash("作业提交成功", "success")
        return redirect(url_for("student.dashboard"))

    flash("提交失败", "danger")
    return redirect(url_for("student.dashboard"))


@student_bp.route("/download/<int:homework_id>")
@login_required
def download(homework_id):
    hw = Homework.query.get_or_404(homework_id)
    # 只有作业所有者可以下载
    if hw.user_id != current_user.id:
        flash('无权访问该文件', 'danger')
        return redirect(url_for('student_dashboard'))

    return send_from_directory(
        current_app.config["UPLOAD_FOLDER"],
        hw.file_path,
        as_attachment=True,
        download_name=hw.file_name,
    )
