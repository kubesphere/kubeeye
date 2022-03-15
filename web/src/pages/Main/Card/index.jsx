/*
 * This file is part of KubeSphere Console.
 * Copyright (C) 2019 The KubeSphere Console Authors.
 *
 * KubeSphere Console is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * KubeSphere Console is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with KubeSphere Console.  If not, see <https://www.gnu.org/licenses/>.
 */

import React, { PureComponent } from 'react'
import PropTypes from 'prop-types'
import classnames from 'classnames'
import { isString } from 'lodash'

import { Loading } from '@kube-design/components'

import styles from './index.scss'

export default class Card extends PureComponent {
  static propTypes = {
    className: PropTypes.string,
    type: PropTypes.string,
    loading: PropTypes.bool,
    refreshing: PropTypes.bool,
    title: PropTypes.oneOfType([
      PropTypes.string,
      PropTypes.element,
      PropTypes.node,
    ]),
    operations: PropTypes.oneOfType([
      PropTypes.string,
      PropTypes.element,
      PropTypes.node,
    ]),
    header: PropTypes.node,
    empty: PropTypes.oneOfType([
      PropTypes.string,
      PropTypes.element,
      PropTypes.node,
    ]),
    isEmpty: PropTypes.bool,
  }

  static defaultProps = {
    title: '',
    type: 'default',
    isEmpty: false,
  }

  renderTitle() {
    const { header, title, operations } = this.props

    if (header) {
      return header
    }

    if (!title && !operations) return null

    return (
      <div className={styles.title}>
        {operations && <div className={styles.operations}>{operations}</div>}
        {title}
      </div>
    )
  }

  renderContent() {
    const { empty, children, isEmpty } = this.props

    if (isEmpty || !children) {
      return isString(empty) ? (
        <div className={styles.empty}>{empty}</div>
      ) : (
        empty
      )
    }

    return children
  }

  render() {
    const {
      className,
      type,
      loading,
      refreshing,
      title,
      operations,
      empty,
      children,
      isEmpty,
      ...rest
    } = this.props

    return (
      <div
        className={classnames(styles.card, className, styles[type])}
        {...rest}
      >
        {this.renderTitle()}
        {loading ? (
          <Loading className={styles.loading} />
        ) : (
          this.renderContent()
        )}
      </div>
    )
  }
}
