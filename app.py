from flask import Flask, redirect, url_for
from config import Config
from extensions import db, login_manager
from routes.auth import auth_bp
from routes.student import student_bp
from routes.admin import admin_bp
import pymysql
import os

# 兼容 MySQLdb
pymysql.install_as_MySQLdb()


def create_app():
    app = Flask(__name__)
    app.config.from_object(Config)
    app.secret_key = Config.SECRET_KEY

    # 创建 uploads 目录
    if not os.path.exists(app.config["UPLOAD_FOLDER"]):
        os.makedirs(app.config["UPLOAD_FOLDER"])

    # 初始化插件
    db.init_app(app)
    login_manager.init_app(app)

    @login_manager.user_loader
    def load_user(user_id):
        return User.query.get(int(user_id))

    @app.route('/')
    def home():
        return redirect(url_for('auth.login'))

    # 初始化数据库（自动建表）
    with app.app_context():
        from models.user import User
        from models.homework import Homework
        db.create_all()

        # 创建默认管理员
        admin = User.query.filter_by(is_admin=True).first()
        if not admin:
            from utils.security import hash_password
            admin = User(
                username='admin',
                student_id='00000000',
                email='admin@example.com',
                password=hash_password('admin'),
                is_admin=True
            )
            db.session.add(admin)
            db.session.commit()

    # 注册蓝图
    app.register_blueprint(auth_bp)
    app.register_blueprint(student_bp)
    app.register_blueprint(admin_bp)

    return app
