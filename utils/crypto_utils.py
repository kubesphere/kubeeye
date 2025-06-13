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
                key_data = f.read()
                # 检查是否需要去除第一行注释 (// filepath:...)
                if key_data.startswith(b'//'):
                    key_lines = key_data.split(b'\n')
                    if len(key_lines) > 1:
                        key_data = key_lines[1].strip()
                
                # 尝试将密钥加载为有效的Fernet密钥
                try:
                    # 验证密钥是否有效
                    Fernet(key_data)
                    logger.info("成功加载加密密钥")
                    return key_data
                except Exception as e:
                    logger.error(f"加载的密钥格式无效，将生成新密钥: {str(e)}")
                    os.rename(KEY_FILE, f"{KEY_FILE}.backup")
        
        # 如果不存在或无效，生成新密钥
        os.makedirs(os.path.dirname(KEY_FILE), exist_ok=True)
        key = Fernet.generate_key()
        with open(KEY_FILE, 'wb') as f:
            f.write(key)
        logger.info("已生成新的加密密钥")
        return key
    except Exception as e:
        logger.error(f"密钥管理错误: {str(e)}", exc_info=True)
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
        
        # 尝试解码并清理可能的格式问题
        try:
            encrypted = base64.urlsafe_b64decode(encrypted_password)
        except Exception as e:
            logger.error(f"Base64解码失败: {str(e)}, 尝试直接解密")
            # 可能密码已经是解密过的，直接返回
            return encrypted_password
            
        # 解密
        try:
            decrypted = f.decrypt(encrypted).decode()
            logger.debug(f"密码解密成功")
            return decrypted
        except Exception as e:
            logger.error(f"Fernet解密失败: {str(e)}")
            # 解密失败，可能密码已经是解密过的，直接返回
            return encrypted_password
    except Exception as e:
        logger.error(f"密码解密过程中发生错误: {str(e)}")
        return encrypted_password  # 解密失败时返回原始加密字符串
