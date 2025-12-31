from flask_login import current_user, login_user
from flask import redirect, url_for, render_template, flash
from flask import Blueprint, render_template, redirect, url_for, flash
from flask_login import login_user, logout_user
from forms.auth import RegisterForm, LoginForm
from models.user import User
from extensions import db
from utils.security import hash_password

auth_bp = Blueprint("auth", __name__)


@auth_bp.route("/login", methods=["GET", "POST"])
def login():
    if current_user.is_authenticated:
        return redirect(url_for("admin.dashboard" if current_user.is_admin else "student.dashboard"))

    form = LoginForm()
    if form.validate_on_submit():
        user: User = User.query.filter_by(
            student_id=form.student_id.data).first()  # type: ignore
        if user and user.password == hash_password(str(form.password.data)):
            login_user(user)
            return redirect(url_for("admin.dashboard" if user.is_admin else "student.dashboard"))
        flash("学号或密码错误", "danger")
    return render_template("login.html", form=form)


@auth_bp.route("/register", methods=["GET", "POST"])
def register():
    form = RegisterForm()
    if form.validate_on_submit():
        if User.query.filter_by(student_id=form.student_id.data).first():
            flash("学号已注册", "danger")
            return redirect(url_for("auth.register"))
        if User.query.filter_by(email=form.email.data).first():
            flash("邮箱已注册", "danger")
            return redirect(url_for("auth.register"))

        user = User(
            username=str(form.username.data),
            student_id=str(form.student_id.data),
            email=str(form.email.data),
            password=hash_password(str(form.password.data)),
        )

        db.session.add(user)
        db.session.commit()
        flash("注册成功，请登录", "success")
        return redirect(url_for("auth.login"))
    return render_template("register.html", form=form)


@auth_bp.route("/logout")
def logout():
    logout_user()
    return redirect(url_for("auth.login"))
