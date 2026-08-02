import { Injectable, Logger } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { Point } from './point.entity';
import { CreatePointDto } from './point.dto';
import { POINT_RATE, BURN_RATE } from './point.constant';
import { logs, SeverityNumber } from '@opentelemetry/api-logs';

const otelLogger = logs.getLogger('point-service');

@Injectable()
export class PointService {
  private readonly logger = new Logger(PointService.name);

  constructor(
    @InjectRepository(Point)
    private pointRepository: Repository<Point>,
  ) {}

  calculatePoint(amount: number): number {
    const point = amount < 0 ? 0 : Math.floor(amount / POINT_RATE);
    this.logger.log(`Point calculated: amount=${amount}, point=${point}`);
    otelLogger.emit({
      severityNumber: SeverityNumber.INFO,
      severityText: 'INFO',
      body: 'Point calculated',
      attributes: {
        log_type: 'business',
        event: 'point_calculated',
        entity_type: 'point',
        amount: amount,
        point: point,
      },
    });
    return point;
  }

  calculateDiscount(
    points: number,
    subtotal: number,
  ): { burn_point: number; discount: number } {
    const safePoints =
      !Number.isFinite(points) || points < 0 ? 0 : Math.floor(points);
    const safeSubtotal =
      !Number.isFinite(subtotal) || subtotal < 0 ? 0 : subtotal;

    // Points burn in pairs (2 points = 1.00 THB); an odd point stays unspent.
    // The discount never exceeds the subtotal - shipping is not discountable.
    const maxByBalance = BURN_RATE * Math.floor(safePoints / BURN_RATE);
    const maxBySubtotal = BURN_RATE * Math.floor(safeSubtotal);
    const burnPoint = Math.min(maxByBalance, maxBySubtotal);
    const discount = burnPoint / BURN_RATE;

    this.logger.log(
      `Discount calculated: points=${points}, subtotal=${subtotal}, burn_point=${burnPoint}, discount=${discount}`,
    );
    otelLogger.emit({
      severityNumber: SeverityNumber.INFO,
      severityText: 'INFO',
      body: 'Discount calculated',
      attributes: {
        log_type: 'business',
        event: 'discount_calculated',
        entity_type: 'point',
        points: points,
        subtotal: subtotal,
        burn_point: burnPoint,
        discount: discount,
      },
    });
    return { burn_point: burnPoint, discount: discount };
  }

  async getPoint(): Promise<Point[]> {
    try {
      const points = await this.pointRepository.find();
      this.logger.log(`Points retrieved, count=${points.length}`);
      otelLogger.emit({
        severityNumber: SeverityNumber.INFO,
        severityText: 'INFO',
        body: 'Points retrieved',
        attributes: {
          log_type: 'business',
          event: 'points_retrieved',
          entity_type: 'point',
          items_count: points.length,
        },
      });
      return points;
    } catch (error) {
      this.logger.error('PointRepository.find internal error', error.stack);
      otelLogger.emit({
        severityNumber: SeverityNumber.ERROR,
        severityText: 'ERROR',
        body: 'PointRepository.find internal error',
        attributes: { 'error.message': error.message },
      });
      throw error;
    }
  }

  async deductPoint(point: CreatePointDto): Promise<Point> {
    try {
      const saved = await this.pointRepository.save(point);
      this.logger.log(
        `Points deducted: userId=${point.userId}, orgId=${point.orgId}, amount=${point.amount}`,
      );
      otelLogger.emit({
        severityNumber: SeverityNumber.INFO,
        severityText: 'INFO',
        body: 'Points deducted',
        attributes: {
          log_type: 'state_change',
          event: 'points_deducted',
          entity_type: 'point',
          entity_id: saved.id,
          changed_by: point.userId,
          org_id: point.orgId,
          amount: point.amount,
        },
      });
      return saved;
    } catch (error) {
      this.logger.error('PointRepository.save internal error', error.stack);
      otelLogger.emit({
        severityNumber: SeverityNumber.ERROR,
        severityText: 'ERROR',
        body: 'PointRepository.save internal error',
        attributes: { 'error.message': error.message },
      });
      throw error;
    }
  }
}
