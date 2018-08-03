import React, {PureComponent} from 'react'
import Authorized, {EDITOR_ROLE} from 'src/auth/Authorized'
import {stripPrefix} from 'src/utils/basepath'

import {Source} from 'src/types'

interface Props {
  source: Source
  currentSource: Source
  toggleWizard: (isVisible: boolean) => (source?: Source) => () => void
}

class ConnectionLink extends PureComponent<Props> {
  public render() {
    const {source, toggleWizard} = this.props
    return (
      <h5 className="margin-zero">
        <Authorized
          requiredRole={EDITOR_ROLE}
          replaceWithIfNotAuthorized={<strong>{source.name}</strong>}
        >
          <span
            onClick={toggleWizard(true)(source)}
            className={`connection-title ${this.className}`}
          >
            <strong>{source.name}</strong>
            {this.default}
          </span>
        </Authorized>
      </h5>
    )
  }

  private get className(): string {
    if (this.isCurrentSource) {
      return 'link-success'
    }

    return ''
  }

  private get default(): string {
    const {source} = this.props
    if (source.default) {
      return ' (Default)'
    }

    return ''
  }

  private get isCurrentSource(): boolean {
    const {source, currentSource} = this.props
    return source.id === currentSource.id
  }
}

export default ConnectionLink
