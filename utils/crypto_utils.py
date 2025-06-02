#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
密码加密工具，用于保护敏感信息
"""

import base64
import os
from cryptography.fernet import Fernet
from cryptography.hazmat.primitives import hashes
from cryptography.hazmat.primitives.kdf.pbkdf2 import PBKDF2HMAC
import logging

logger = logging.getLogger(__name__)

# 生成一个默认的密钥文件路径
KEY_FILE = os.path.join(os.path.dirname(os.path.dirname(__file__)), 'data', '.secret_key')

def get_encryption_key():
    """
    获取或生成加密密钥
    
    Returns:
        bytes: 加密密钥
    """
    try:
        # 尝试读取现有密钥
        if os.path.exists(KEY_FILE):
            with open(KEY_FILE, 'rb') as f:
                return f.read()
        
        # 如果不存在，生成新密钥
        os.makedirs(os.path.dirname(KEY_FILE), exist_ok=True)
        key = Fernet.generate_key()
        with open(KEY_FILE, 'wb') as f:
            f.write(key)
        return key
    except Exception as e:
        logger.error(f"密钥管理错误: {str(e)}")
        # 如果出现错误，生成临时密钥（不保存）
        return Fernet.generate_key()

def encrypt_password(password: str) -> str:
    """
    加密密码
    
    Args:
        password: 明文密码
    
    Returns:
        str: 加密后的密码字符串
    """
    if not password:
        return ""
        
    try:
        key = get_encryption_key()
        f = Fernet(key)
        encrypted = f.encrypt(password.encode())
        return base64.urlsafe_b64encode(encrypted).decode()
    except Exception as e:
        logger.error(f"密码加密错误: {str(e)}")
        return password  # 加密失败时返回原始密码

def decrypt_password(encrypted_password: str) -> str:
    """
    解密密码
    
    Args:
        encrypted_password: 加密后的密码字符串
    
    Returns:
        str: 解密后的明文密码
    """
    if not encrypted_password:
        return ""
        
    try:
        key = get_encryption_key()
        f = Fernet(key)
        encrypted = base64.urlsafe_b64decode(encrypted_password)
        return f.decrypt(encrypted).decode()
    except Exception as e:
        logger.error(f"密码解密错误: {str(e)}")
        return encrypted_password  # 解密失败时返回原始加密字符串
