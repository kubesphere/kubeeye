#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
版本信息模块，用于管理项目版本信息
"""

# 主版本号
VERSION_MAJOR = 1
# 次版本号
VERSION_MINOR = 0
# 修订号
VERSION_PATCH = 0
# 版本标签（如 'alpha'、'beta'、'rc1'，正式版留空）
VERSION_TAG = ''

# 完整版本号
VERSION = f"{VERSION_MAJOR}.{VERSION_MINOR}.{VERSION_PATCH}"
if VERSION_TAG:
    VERSION = f"{VERSION}-{VERSION_TAG}"

# 应用名称
APP_NAME = "kubeeye"
# 应用描述
APP_DESCRIPTION = "Kubernetes 集群巡检工具"
# 应用作者
APP_AUTHOR = "pixiake"
# 应用主页
APP_URL = "https://github.com/pixiake/kubeeye"

# 版本发布日期
RELEASE_DATE = "2025-05-28"

def get_version():
    """获取当前版本号"""
    return VERSION

# 版本信息字典
VERSION_INFO = {
    'name': APP_NAME,
    'version': VERSION,
    'description': APP_DESCRIPTION,
    'author': APP_AUTHOR,
    'url': APP_URL,
    'release_date': RELEASE_DATE
}

def get_version_info():
    """获取版本信息字典"""
    return VERSION_INFO

def get_version_string():
    """获取版本字符串"""
    return f"{APP_NAME} v{VERSION}"

if __name__ == "__main__":
    print(get_version_string())
